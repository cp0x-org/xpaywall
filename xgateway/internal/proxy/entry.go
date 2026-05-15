package proxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tempoxyz/mpp-go/pkg/mpp"
	mppserver "github.com/tempoxyz/mpp-go/pkg/server"
	x402 "github.com/x402-foundation/x402/go"
	x402http "github.com/x402-foundation/x402/go/http"
	evmexact "github.com/x402-foundation/x402/go/mechanisms/evm/exact/server"
	evmupto "github.com/x402-foundation/x402/go/mechanisms/evm/upto/server"

	"github.com/cp0x-org/xpaywall/xgateway/internal/config"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

// entry is the per-route resource bundle resolved on first request and cached.
// It owns the upstream reverse proxy plus any payment-protocol handlers.
type entry struct {
	rule    rules.Rule
	rp      *httputil.ReverseProxy
	x402Srv *x402http.HTTPServer
	mppCfgs []mppserver.ComposeConfig
	scheme  string
}

// hasPayment reports whether any payment protocol is configured for this entry.
func (e *entry) hasPayment() bool {
	return e.x402Srv != nil || len(e.mppCfgs) > 0
}

// runPayment dispatches to the appropriate protocol middleware based on what
// the entry has configured and what the client signaled in the request.
// Returns the protocol name used ("x402" or "mpp") for logging; an empty
// string when no protocol fired (should not happen — hasPayment guards this).
//
// Selection rules (mirrors the original switch):
//   - MPP only        → MPP
//   - x402 only       → x402
//   - both, MPP auth  → MPP
//   - both, otherwise → x402 (response is decorated with MPP challenges so the
//     client can opt into MPP on the next request).
func (e *entry) runPayment(c *gin.Context) string {
	hasX402 := e.x402Srv != nil
	hasMPP := len(e.mppCfgs) > 0

	switch {
	case hasMPP && !hasX402:
		runMPPMiddleware(c, e.mppCfgs)
		return "mpp"
	case hasX402 && !hasMPP:
		runX402Middleware(c, e.x402Srv)
		return "x402"
	case hasX402 && hasMPP:
		if mpp.FindPaymentAuthorization(c.GetHeader("Authorization")) != "" {
			runMPPMiddleware(c, e.mppCfgs)
			return "mpp"
		}
		runX402Middleware(c, e.x402Srv, func(headers http.Header) {
			appendMPPChallenges(c.Request.Context(), headers, e.mppCfgs)
		})
		return "x402"
	}
	return ""
}

// buildEntry constructs and initializes the per-route entry: upstream proxy
// (if a target is configured) plus payment-protocol handlers for every enabled
// channel.
func (s *state) buildEntry(ctx context.Context, rule *rules.Rule, reqPath string) (*entry, error) {
	log.Printf("[entry] building entry for %s: free=%v channels=%d", reqPath, rule.Free, len(rule.PaymentChannels))
	e := &entry{rule: *rule}

	if rule.Target != "" {
		rp, err := buildUpstreamProxy(rule, reqPath)
		if err != nil {
			return nil, err
		}
		e.rp = rp
	}

	if rule.Free || len(rule.PaymentChannels) == 0 {
		return e, nil
	}

	x402Options, mppCfgs, scheme, err := compilePaymentChannels(rule)
	if err != nil {
		return nil, err
	}
	e.scheme = scheme
	e.mppCfgs = mppCfgs

	if len(x402Options) == 0 {
		return e, nil
	}

	srv, err := s.buildX402Server(ctx, rule, reqPath, x402Options)
	if err != nil {
		return nil, err
	}
	e.x402Srv = srv
	return e, nil
}

// buildUpstreamProxy creates the reverse proxy for the rule's Target. It strips
// the project-slug prefix before forwarding, injects the configured auth header
// (when set), and rewrites Location headers on upstream redirects so that the
// client keeps the slug prefix.
func buildUpstreamProxy(rule *rules.Rule, reqPath string) (*httputil.ReverseProxy, error) {
	upURL, err := url.Parse(rule.Target)
	if err != nil {
		return nil, fmt.Errorf("parse target URL: %w", err)
	}
	authName := rule.AuthHeaderName
	authValue := rule.AuthHeaderValue

	// cleanPath is the path the upstream expects (project slug stripped).
	cleanPath := rule.InboundPath

	// Compute the project-slug prefix so we can rewrite upstream redirect
	// Location headers that strip it (e.g. /health → /project/health).
	prefix := ""
	if cleanPath != "" && strings.HasSuffix(reqPath, cleanPath) {
		prefix = reqPath[:len(reqPath)-len(cleanPath)]
	}

	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = upURL.Scheme
			r.URL.Host = upURL.Host
			r.Host = upURL.Host
			// Strip project-slug prefix before forwarding to upstream.
			if cleanPath != "" && cleanPath != r.URL.Path && !containsGlob(cleanPath) {
				r.URL.Path = cleanPath
			}
			if authName != "" && authValue != "" {
				r.Header.Set(authName, authValue)
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			if prefix != "" && resp.StatusCode >= 300 && resp.StatusCode < 400 {
				if loc := resp.Header.Get("Location"); strings.HasPrefix(loc, "/") && !strings.HasPrefix(loc, prefix+"/") {
					resp.Header.Set("Location", prefix+loc)
				}
			}
			scheme, _ := resp.Request.Context().Value(matchedRuleSchemeContextKey).(string)
			applyUpstreamSettlementHeaders(resp.Header, scheme)
			return nil
		},
	}, nil
}

// compilePaymentChannels iterates over the rule's enabled channels and produces
// per-protocol data structures: a list of x402 PaymentOptions, a list of MPP
// compose configs, and the entry's effective scheme (taken from the first
// channel that contributes one). Channel-level price overrides the route price.
func compilePaymentChannels(rule *rules.Rule) (
	x402Options x402http.PaymentOptions,
	mppCfgs []mppserver.ComposeConfig,
	scheme string,
	err error,
) {
	x402Options = make(x402http.PaymentOptions, 0)

	for _, ch := range rule.PaymentChannels {
		if !ch.Enabled {
			continue
		}
		price := ch.Price
		if price == "" {
			price = rule.Price
		}
		switch ch.Protocol {
		case "x402":
			x402Options = append(x402Options, x402http.PaymentOption{
				Scheme:  config.NormalizeScheme(ch.Scheme),
				PayTo:   ch.ChannelConfig["merchant"],
				Price:   price,
				Network: x402.Network(ch.ChannelConfig["network"]),
			})
			if scheme == "" {
				scheme = config.NormalizeScheme(ch.Scheme)
			}
		case "mpp":
			mppSrv, err := buildMPPServer(config.MPPMethod{
				Method:    ch.Method,
				Scheme:    ch.Scheme,
				RPCURL:    ch.ChannelConfig["rpc_url"],
				Merchant:  ch.ChannelConfig["merchant"],
				Asset:     ch.ChannelConfig["asset"],
				SecretKey: ch.ChannelConfig["secret_key"],
			})
			if err != nil {
				return nil, nil, "", fmt.Errorf("build MPP server: %w", err)
			}
			mppCfgs = append(mppCfgs, mppserver.ComposeConfig{
				Mpp: mppSrv,
				Params: mppserver.ChargeParams{
					Amount:      parseMPPAmount(price),
					Currency:    ch.ChannelConfig["asset"],
					Recipient:   ch.ChannelConfig["merchant"],
					Description: rule.Description,
				},
			})
			if scheme == "" {
				scheme = config.NormalizeScheme(ch.Scheme)
			}
		}
	}
	return x402Options, mppCfgs, scheme, nil
}

// buildX402Server initializes the x402 HTTP server for a rule's x402 channels:
// attaches each channel's facilitator client (via state.facilitator cache),
// registers EVM exact/upto scheme handlers, and runs the library's Initialize
// step under a 30s timeout.
//
// Channels missing a facilitator_url are skipped with a warning — verification
// will fail later but route construction does not abort.
func (s *state) buildX402Server(
	ctx context.Context,
	rule *rules.Rule,
	reqPath string,
	options x402http.PaymentOptions,
) (*x402http.HTTPServer, error) {
	// Always use the full proxy path (with project slug) as the x402 route key.
	// The payment challenge must reference the same URL the client originally requested
	// so the retry after payment hits the proxy at the correct slug-bearing path.
	routeKey := reqPath
	log.Printf("[entry] x402 routeKey=%s options=%d", routeKey, len(options))

	routes := x402http.RoutesConfig{
		routeKey: x402http.RouteConfig{
			Accepts:     options,
			Description: rule.Description,
			MimeType:    rule.MimeType,
		},
	}

	var x402Opts []x402.ResourceServerOption
	for _, ch := range rule.PaymentChannels {
		if ch.Protocol != "x402" || !ch.Enabled {
			continue
		}
		facURL := ch.ChannelConfig["facilitator_url"]
		log.Printf("[entry] x402 channel scheme=%s network=%s facilitator_url=%q merchant=%q",
			ch.Scheme, ch.ChannelConfig["network"], facURL, ch.ChannelConfig["merchant"])
		if facURL == "" {
			log.Printf("[entry] WARNING: x402 channel missing facilitator_url — payment verification will fail")
			continue
		}
		x402Opts = append(x402Opts, x402.WithFacilitatorClient(s.facilitator(facURL)))
	}

	srv := x402http.Newx402HTTPResourceServer(routes, x402Opts...)
	for _, ch := range rule.PaymentChannels {
		if ch.Protocol != "x402" || !ch.Enabled {
			continue
		}
		network := x402.Network(ch.ChannelConfig["network"])
		switch config.NormalizeScheme(ch.Scheme) {
		case config.SchemeExact:
			srv.Register(network, evmexact.NewExactEvmScheme())
		case config.SchemeUpto:
			srv.Register(network, evmupto.NewUptoEvmScheme())
		}
	}

	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := srv.Initialize(initCtx); err != nil {
		return nil, fmt.Errorf("initialize x402 server: %w", err)
	}
	return srv, nil
}

// classifyUpstreamResult maps a (status, err) pair from the reverse proxy into
// the log entry's final-status, error-type, error-message triple.
func classifyUpstreamResult(statusCode int32, err error) (proxyStatus, string, string) {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return statusUpstreamTimeout, "timeout", err.Error()
		}
		return statusFailed, "proxy_error", err.Error()
	}
	switch {
	case statusCode == http.StatusGatewayTimeout:
		return statusUpstreamTimeout, "timeout", ""
	case statusCode == http.StatusBadGateway || statusCode == http.StatusServiceUnavailable:
		return statusUpstreamError, "upstream_unavailable", ""
	case statusCode >= 500:
		return statusUpstreamError, "upstream_error", ""
	default:
		return statusCompleted, "", ""
	}
}

// resolvePaymentChannelInfo finds the first enabled channel matching usedProtocol
// and returns its DB IDs and the effective price as amountUSD.
// Channel-level price takes precedence over routePrice.
// Zero UUIDs (e.g. from file provider) are treated as "unknown" and returned as nil
// to avoid FK violations in the database.
func resolvePaymentChannelInfo(channels []*rules.PaymentChannel, usedProtocol, routePrice string) (*uuid.UUID, *uuid.UUID, *string) {
	zeroUUID := uuid.UUID{}
	for _, ch := range channels {
		if ch.Protocol != usedProtocol || !ch.Enabled {
			continue
		}

		// Resolve the effective price: channel-specific overrides route-level.
		price := ch.Price
		if price == "" {
			price = routePrice
		}
		var pricePtr *string
		if price != "" {
			pricePtr = &price
		}

		// Only return channel/asset IDs when they are real DB UUIDs.
		var channelID, assetID *uuid.UUID
		if ch.ID != zeroUUID {
			id := ch.ID
			channelID = &id
		}
		if ch.AssetID != nil && *ch.AssetID != zeroUUID {
			assetID = ch.AssetID
		}

		return channelID, assetID, pricePtr
	}
	return nil, nil, nil
}

func containsGlob(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func strPtr(s string) *string { return &s }
func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

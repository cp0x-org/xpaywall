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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cp0x-org/xpaywall/xgateway/internal/config"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

// entry is the per-route resource bundle resolved on first request and cached.
// It owns the upstream reverse proxy plus any payment protocols configured for
// the route. Protocols are stored in the Server's registry order so that the
// first entry acts as the default fallback.
type entry struct {
	rule      rules.Rule
	rp        *httputil.ReverseProxy
	protocols []PaymentProtocol
	scheme    string
}

// hasPayment reports whether any payment protocol is configured for this entry.
func (e *entry) hasPayment() bool {
	return len(e.protocols) > 0
}

// runPayment dispatches the request to one of the entry's protocols. Selection:
//   - if any protocol's HasClientAuth reports true, that protocol handles the
//     request; the other protocols contribute Challenge decorators so the 402
//     response (if any) still advertises them.
//   - otherwise the first registered protocol handles the request, with every
//     other protocol contributing Challenge decorators.
//
// Returns the protocol name used for logging; empty when no protocol is
// configured (caller guards via hasPayment).
func (e *entry) runPayment(c *gin.Context) string {
	if len(e.protocols) == 0 {
		return ""
	}

	chosen := 0
	for i, p := range e.protocols {
		if p.HasClientAuth(c) {
			chosen = i
			break
		}
	}

	decorators := make([]func(http.Header), 0, len(e.protocols)-1)
	for i, p := range e.protocols {
		if i == chosen {
			continue
		}
		other := p
		decorators = append(decorators, func(h http.Header) {
			other.Challenge(c.Request.Context(), h)
		})
	}

	e.protocols[chosen].Handle(c, decorators...)
	return e.protocols[chosen].Name()
}

// buildEntry constructs and initializes the per-route entry: upstream proxy
// (if a target is configured) plus payment protocols built from the rule's
// enabled channels grouped by protocol name. Protocols are instantiated in the
// Server's registry order; a protocol factory that returns (nil, nil) is
// skipped without aborting entry construction.
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

	// Effective scheme is taken from the first enabled channel — matches the
	// previous behavior of compilePaymentChannels.
	for _, ch := range rule.PaymentChannels {
		if !ch.Enabled {
			continue
		}
		e.scheme = config.NormalizeScheme(ch.Scheme)
		break
	}

	byProto := groupEnabledChannelsByProtocol(rule.PaymentChannels)
	for _, reg := range s.protocols {
		channels := byProto[reg.name]
		if len(channels) == 0 {
			continue
		}
		p, err := reg.factory(ctx, rule, reqPath, channels)
		if err != nil {
			return nil, fmt.Errorf("build %s protocol: %w", reg.name, err)
		}
		if p == nil {
			continue
		}
		e.protocols = append(e.protocols, p)
	}

	return e, nil
}

// groupEnabledChannelsByProtocol bucketises the rule's enabled channels by
// their Protocol field while preserving the original channel order within each
// bucket.
func groupEnabledChannelsByProtocol(channels []*rules.PaymentChannel) map[string][]*rules.PaymentChannel {
	byProto := make(map[string][]*rules.PaymentChannel)
	for _, ch := range channels {
		if !ch.Enabled {
			continue
		}
		byProto[ch.Protocol] = append(byProto[ch.Protocol], ch)
	}
	return byProto
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

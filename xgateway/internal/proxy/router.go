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
	"sync"
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
	"github.com/cp0x-org/xpaywall/xgateway/internal/logger"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

const ctxEntryKey = "proxy.entry"

type entry struct {
	rule    rules.Rule
	rp      *httputil.ReverseProxy
	x402Srv *x402http.HTTPServer
	mppCfgs []mppserver.ComposeConfig
	scheme  string
}

// pendingLogEntry tracks a 402-response log entry waiting for its payment retry.
type pendingLogEntry struct {
	logID     uuid.UUID
	expiresAt time.Time
}

type state struct {
	provider    rules.Provider
	fallback    http.Handler
	logger      *logger.Client
	startedAt   time.Time
	mu          sync.RWMutex
	routeCache  map[string]*entry
	facilCache  map[string]x402.FacilitatorClient
	pendingLogs sync.Map // key: logFingerprint → pendingLogEntry
}

func newState(provider rules.Provider, fallback http.Handler, lg *logger.Client) *state {
	return &state{
		provider:   provider,
		fallback:   fallback,
		logger:     lg,
		startedAt:  time.Now(),
		routeCache: make(map[string]*entry),
		facilCache: make(map[string]x402.FacilitatorClient),
	}
}

// logFingerprint returns a key that correlates the initial 402 request with its
// payment retry. Uses method+path+clientIP — reliable for single-client flows.
func logFingerprint(method, path, clientIP string) string {
	return method + "|" + path + "|" + clientIP
}

func (s *state) buildRouter() http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Any("/*path", s.loggingMiddleware(), s.resolve(), s.proxyUpstream())

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthHandler())
	mux.Handle("/", router)
	return mux
}

func (s *state) healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","started_at":%q,"uptime_seconds":%d}`,
			s.startedAt.UTC().Format(time.RFC3339),
			int64(time.Since(s.startedAt).Seconds()),
		)
	}
}

// loggingMiddleware wraps the entire request: sets up logContext, runs handlers,
// then dispatches collected data to the logger in a background goroutine.
//
// HEAD requests are skipped entirely — browsers send them automatically and
// they carry no useful information for payment tracking.
//
// For payment-required routes the middleware maintains a short-lived in-memory
// map keyed by (method+path+clientIP). When a 402 is returned the entry is
// stored so the subsequent paid retry can UPDATE the same DB row instead of
// creating a duplicate.
func (s *state) loggingMiddleware() gin.HandlerFunc {
	const pendingLogTTL = 10 * time.Minute

	return func(c *gin.Context) {
		// Browser HEAD requests are noise — skip entirely.
		if c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		lc := newLogContext()

		// Check if this is a retry for a previously-recorded 402.
		fp := logFingerprint(c.Request.Method, c.Request.URL.Path, c.ClientIP())
		if v, ok := s.pendingLogs.Load(fp); ok {
			pe := v.(pendingLogEntry)
			if time.Now().Before(pe.expiresAt) {
				lc.logID = pe.logID
				lc.isRetry = true
			} else {
				s.pendingLogs.Delete(fp)
			}
		}

		c.Set(ctxLogKey, lc)
		c.Next()

		// Update the pending-log cache based on outcome.
		if lc.ruleFound {
			switch lc.paymentEvent {
			case statusPaymentRequired:
				// Store so the next retry can find this log ID.
				s.pendingLogs.Store(fp, pendingLogEntry{
					logID:     lc.logID,
					expiresAt: time.Now().Add(pendingLogTTL),
				})
			default:
				// Request completed (paid or free) — remove pending entry.
				s.pendingLogs.Delete(fp)
			}
		}

		lc.dispatch(s.logger)
	}
}

func (s *state) resolve() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path

		s.mu.RLock()
		e, ok := s.routeCache[reqPath]
		s.mu.RUnlock()

		if !ok {
			rule, err := s.provider.GetByInboundPath(c.Request.Context(), reqPath)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "resolve failed: " + err.Error()})
				c.Abort()
				return
			}
			if rule == nil {
				if s.fallback != nil {
					s.fallback.ServeHTTP(c.Writer, c.Request)
					c.Abort()
					return
				}
				c.JSON(http.StatusNotFound, gin.H{
					"error":  "route not found",
					"path":   reqPath,
					"status": http.StatusNotFound,
				})
				c.Abort()
				return
			}

			e, err = s.buildEntry(c.Request.Context(), rule, reqPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "payment setup failed: " + err.Error()})
				c.Abort()
				return
			}

			s.mu.Lock()
			s.routeCache[reqPath] = e
			s.mu.Unlock()
		}

		// Populate log context now that we have the rule.
		if lc := getLogContext(c); lc != nil {
			clientIP := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")
			lc.setRule(&e.rule, c.Request.Method, reqPath, strPtr(clientIP), strPtrIfNotEmpty(userAgent))
		}

		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), matchedRuleSchemeContextKey, e.scheme),
		)
		c.Set(ctxEntryKey, e)

		if e.rule.Free || (e.x402Srv == nil && len(e.mppCfgs) == 0) {
			c.Next()
			return
		}

		hasPaymentAuth := mpp.FindPaymentAuthorization(c.GetHeader("Authorization")) != ""
		var usedProtocol string
		switch {
		case len(e.mppCfgs) > 0 && e.x402Srv == nil:
			runMPPMiddleware(c, e.mppCfgs)
			usedProtocol = "mpp"
		case e.x402Srv != nil && len(e.mppCfgs) == 0:
			runX402Middleware(c, e.x402Srv)
			usedProtocol = "x402"
		case e.x402Srv != nil && len(e.mppCfgs) > 0:
			if hasPaymentAuth {
				runMPPMiddleware(c, e.mppCfgs)
				usedProtocol = "mpp"
			} else {
				runX402Middleware(c, e.x402Srv, func(headers http.Header) {
					appendMPPChallenges(c.Request.Context(), headers, e.mppCfgs)
				})
				usedProtocol = "x402"
			}
		}

		// Update log context with payment outcome.
		if lc := getLogContext(c); lc != nil {
			if c.IsAborted() {
				lc.setPaymentRequired()
			} else {
				channelID, assetID, amountUSD := resolvePaymentChannelInfo(e.rule.PaymentChannels, usedProtocol, e.rule.Price)
				lc.setPaymentCompleted(channelID, assetID, amountUSD)
			}
		}
	}
}

func (s *state) proxyUpstream() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(ctxEntryKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing route entry"})
			return
		}
		e := raw.(*entry)
		if e.rp == nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "no upstream configured"})
			return
		}

		lc := getLogContext(c)
		if lc != nil {
			lc.setProxying(e.rule.Target)
		}

		start := time.Now()
		e.rp.ServeHTTP(c.Writer, c.Request)
		elapsedMs := int32(time.Since(start).Milliseconds())
		statusCode := int32(c.Writer.Status())

		if lc != nil {
			finalStatus, errType, errMsg := classifyUpstreamResult(statusCode, nil)
			lc.setUpstreamResult(statusCode, elapsedMs, finalStatus, errType, errMsg)
		}
	}
}

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

func (s *state) buildEntry(ctx context.Context, rule *rules.Rule, reqPath string) (*entry, error) {
	log.Printf("[entry] building entry for %s: free=%v channels=%d", reqPath, rule.Free, len(rule.PaymentChannels))
	e := &entry{rule: *rule}

	if rule.Target != "" {
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

		e.rp = &httputil.ReverseProxy{
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
		}
	}

	if rule.Free || len(rule.PaymentChannels) == 0 {
		return e, nil
	}

	x402Options := make(x402http.PaymentOptions, 0)
	var mppCfgs []mppserver.ComposeConfig

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
			if e.scheme == "" {
				e.scheme = config.NormalizeScheme(ch.Scheme)
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
				return nil, fmt.Errorf("build MPP server: %w", err)
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
			if e.scheme == "" {
				e.scheme = config.NormalizeScheme(ch.Scheme)
			}
		}
	}

	e.mppCfgs = mppCfgs

	if len(x402Options) == 0 {
		return e, nil
	}

	// Always use the full proxy path (with project slug) as the x402 route key.
	// The payment challenge must reference the same URL the client originally requested
	// so the retry after payment hits the proxy at the correct slug-bearing path.
	routeKey := reqPath
	log.Printf("[entry] x402 routeKey=%s options=%d", routeKey, len(x402Options))

	routes := x402http.RoutesConfig{
		routeKey: x402http.RouteConfig{
			Accepts:     x402Options,
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
		s.mu.RLock()
		fac, exists := s.facilCache[facURL]
		s.mu.RUnlock()
		if !exists {
			fac = x402http.NewHTTPFacilitatorClient(&x402http.FacilitatorConfig{
				URL:        facURL,
				HTTPClient: FacilitatorHTTPClient(),
			})
			s.mu.Lock()
			s.facilCache[facURL] = fac
			s.mu.Unlock()
		}
		x402Opts = append(x402Opts, x402.WithFacilitatorClient(fac))
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

	e.x402Srv = srv
	return e, nil
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

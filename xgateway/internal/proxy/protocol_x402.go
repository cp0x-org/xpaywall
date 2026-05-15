package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	x402 "github.com/x402-foundation/x402/go"
	x402http "github.com/x402-foundation/x402/go/http"
	evmexact "github.com/x402-foundation/x402/go/mechanisms/evm/exact/server"
	evmupto "github.com/x402-foundation/x402/go/mechanisms/evm/upto/server"

	"github.com/cp0x-org/xpaywall/xgateway/internal/config"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

const (
	protoX402                   = "x402"
	upstreamSettlementAmountKey = "X-X402-Settlement-Amount"
	x402DefaultTimeout          = 30 * time.Second
)

type requestContextKey string

const matchedRuleSchemeContextKey requestContextKey = "proxy.rule_scheme"

// ---------- Facilitator HTTP client (used by every x402 protocol instance) --

// facilitatorLoggingTransport wraps http.DefaultTransport and logs every
// request sent to the facilitator and the full response received.
type facilitatorLoggingTransport struct{}

func (t *facilitatorLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}
	log.Printf("[facilitator] → %s %s body=%s", req.Method, req.URL, string(reqBody))

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("[facilitator] ← error: %v", err)
		return nil, err
	}

	respBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	log.Printf("[facilitator] ← %d %s body=%s", resp.StatusCode, req.URL, string(respBody))
	return resp, nil
}

// FacilitatorHTTPClient returns an http.Client with facilitator request/response logging.
func FacilitatorHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   x402DefaultTimeout,
		Transport: &facilitatorLoggingTransport{},
	}
}

// facilitatorCache memoises x402 facilitator clients across all routes within
// a single Server. The cache key is the facilitator URL.
type facilitatorCache struct {
	mu sync.RWMutex
	m  map[string]x402.FacilitatorClient
}

func newFacilitatorCache() *facilitatorCache {
	return &facilitatorCache{m: make(map[string]x402.FacilitatorClient)}
}

func (f *facilitatorCache) Get(url string) x402.FacilitatorClient {
	f.mu.RLock()
	fac, ok := f.m[url]
	f.mu.RUnlock()
	if ok {
		return fac
	}
	fac = x402http.NewHTTPFacilitatorClient(&x402http.FacilitatorConfig{
		URL:        url,
		HTTPClient: FacilitatorHTTPClient(),
	})
	f.mu.Lock()
	f.m[url] = fac
	f.mu.Unlock()
	return fac
}

// ---------- x402 protocol implementation ------------------------------------

// x402Protocol is a PaymentProtocol backed by an initialized x402 HTTP server
// covering one route's enabled x402 channels.
type x402Protocol struct {
	srv *x402http.HTTPServer
}

func (p *x402Protocol) Name() string { return protoX402 }

// HasClientAuth: x402 is the registry's default protocol. Selection rolls back
// to the first registered protocol when no other protocol's HasClientAuth
// matches, so x402 does not need to advertise its own positive signal here.
func (p *x402Protocol) HasClientAuth(_ *gin.Context) bool { return false }

func (p *x402Protocol) Handle(c *gin.Context, decorators ...func(http.Header)) {
	if p.srv == nil {
		c.Next()
		return
	}

	adapter := &ginAdapter{ctx: c}
	reqCtx := x402http.HTTPRequestContext{
		Adapter: adapter,
		Path:    c.Request.URL.Path,
		Method:  c.Request.Method,
	}

	if !p.srv.RequiresPayment(reqCtx) {
		c.Next()
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), x402DefaultTimeout)
	defer cancel()

	result := p.srv.ProcessHTTPRequest(ctx, reqCtx, nil)
	log.Printf("[x402] ProcessHTTPRequest path=%s result=%v", reqCtx.Path, result.Type)
	switch result.Type {
	case x402http.ResultNoPaymentRequired:
		c.Next()
	case x402http.ResultPaymentError:
		if result.Response != nil {
			log.Printf("[x402] payment error: status=%d body=%v", result.Response.Status, result.Response.Body)
		}
		writeX402Error(c, result.Response, decorators...)
	case x402http.ResultPaymentVerified:
		log.Printf("[x402] payment verified, proceeding to upstream")
		handleX402Verified(c, p.srv, ctx, reqCtx, result)
	}
}

// Challenge: x402 expresses its requirements via the full 402 response body
// (JSON with payment requirements), not via standalone headers — so this is a
// no-op. Other protocols ignore x402 as a fallback challenge.
func (p *x402Protocol) Challenge(_ context.Context, _ http.Header) {}

// ---------- x402 factory ----------------------------------------------------

// newX402Factory returns a ProtocolFactory bound to the given facilitator
// cache. Each Server has its own cache (shared across that Server's routes).
func newX402Factory(cache *facilitatorCache) ProtocolFactory {
	return func(ctx context.Context, rule *rules.Rule, reqPath string, channels []*rules.PaymentChannel) (PaymentProtocol, error) {
		options := make(x402http.PaymentOptions, 0, len(channels))
		for _, ch := range channels {
			price := ch.Price
			if price == "" {
				price = rule.Price
			}
			options = append(options, x402http.PaymentOption{
				Scheme:  config.NormalizeScheme(ch.Scheme),
				PayTo:   ch.ChannelConfig["merchant"],
				Price:   price,
				Network: x402.Network(ch.ChannelConfig["network"]),
			})
		}
		if len(options) == 0 {
			return nil, nil
		}

		// Always use the full proxy path (with project slug) as the x402 route key.
		// The payment challenge must reference the same URL the client originally
		// requested so that the retry after payment hits the proxy at the correct
		// slug-bearing path.
		routeKey := reqPath
		log.Printf("[entry] x402 routeKey=%s options=%d", routeKey, len(options))

		routes := x402http.RoutesConfig{
			routeKey: x402http.RouteConfig{
				Accepts:     options,
				Description: rule.Description,
				MimeType:    rule.MimeType,
			},
		}

		var serverOpts []x402.ResourceServerOption
		for _, ch := range channels {
			facURL := ch.ChannelConfig["facilitator_url"]
			log.Printf("[entry] x402 channel scheme=%s network=%s facilitator_url=%q merchant=%q",
				ch.Scheme, ch.ChannelConfig["network"], facURL, ch.ChannelConfig["merchant"])
			if facURL == "" {
				log.Printf("[entry] WARNING: x402 channel missing facilitator_url — payment verification will fail")
				continue
			}
			serverOpts = append(serverOpts, x402.WithFacilitatorClient(cache.Get(facURL)))
		}

		srv := x402http.Newx402HTTPResourceServer(routes, serverOpts...)
		for _, ch := range channels {
			network := x402.Network(ch.ChannelConfig["network"])
			switch config.NormalizeScheme(ch.Scheme) {
			case config.SchemeExact:
				srv.Register(network, evmexact.NewExactEvmScheme())
			case config.SchemeUpto:
				srv.Register(network, evmupto.NewUptoEvmScheme())
			}
		}

		initCtx, cancel := context.WithTimeout(ctx, x402DefaultTimeout)
		defer cancel()
		if err := srv.Initialize(initCtx); err != nil {
			return nil, fmt.Errorf("initialize x402 server: %w", err)
		}
		return &x402Protocol{srv: srv}, nil
	}
}

// ---------- x402 runtime helpers --------------------------------------------

func writeX402Error(c *gin.Context, response *x402http.HTTPResponseInstructions, decorateHeaders ...func(http.Header)) {
	if response == nil {
		c.AbortWithStatus(http.StatusPaymentRequired)
		return
	}
	for key, value := range response.Headers {
		c.Header(key, value)
	}
	for _, decorate := range decorateHeaders {
		if decorate != nil {
			decorate(c.Writer.Header())
		}
	}
	c.Status(response.Status)
	if response.IsHTML {
		c.Data(response.Status, "text/html; charset=utf-8", []byte(fmt.Sprint(response.Body)))
	} else {
		c.JSON(response.Status, response.Body)
	}
	c.Abort()
}

func handleX402Verified(c *gin.Context, server *x402http.HTTPServer, ctx context.Context, reqCtx x402http.HTTPRequestContext, result x402http.HTTPProcessResult) {
	writer := &responseCapture{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
	c.Writer = writer

	c.Next()
	if c.IsAborted() {
		return
	}

	c.Writer = writer.ResponseWriter
	if writer.statusCode >= 400 {
		c.Writer.WriteHeader(writer.statusCode)
		_, _ = c.Writer.Write(writer.body.Bytes())
		return
	}

	log.Printf("[x402] running settlement, upstream status=%d", writer.statusCode)
	settleResult := server.ProcessSettlement(
		ctx,
		*result.PaymentPayload,
		*result.PaymentRequirements,
		nil,
		&x402http.HTTPTransportContext{
			Request:         &reqCtx,
			ResponseBody:    writer.body.Bytes(),
			ResponseHeaders: writer.Header(),
		},
	)

	if !settleResult.Success {
		log.Printf("[x402] settlement FAILED: headers=%v response=%v", settleResult.Headers, settleResult.Response)
		for key, value := range settleResult.Headers {
			c.Header(key, value)
		}
		if settleResult.Response != nil {
			writeX402Error(c, settleResult.Response)
			return
		}
		c.AbortWithStatus(http.StatusPaymentRequired)
		return
	}
	log.Printf("[x402] settlement SUCCESS")

	for key, value := range settleResult.Headers {
		c.Header(key, value)
	}
	c.Writer.WriteHeader(writer.statusCode)
	_, _ = c.Writer.Write(writer.body.Bytes())
}

// applyUpstreamSettlementHeaders translates the upstream's "actual amount used"
// signal (X-X402-Settlement-Amount header) into the x402 SettlementOverrides
// header that the x402 library understands. Only the x402 "upto" scheme needs
// this dance; for any other scheme the helper just strips the headers.
//
// Called from the reverse-proxy ModifyResponse hook in buildUpstreamProxy and
// from NewReverseProxy.
func applyUpstreamSettlementHeaders(headers http.Header, scheme string) {
	if headers == nil {
		return
	}
	if config.NormalizeScheme(scheme) != config.SchemeUpto {
		headers.Del(x402http.SettlementOverridesHeader)
		headers.Del(upstreamSettlementAmountKey)
		return
	}
	amount := strings.TrimSpace(headers.Get(upstreamSettlementAmountKey))
	if amount != "" && headers.Get(x402http.SettlementOverridesHeader) == "" {
		headers.Set(
			x402http.SettlementOverridesHeader,
			x402http.MarshalSettlementOverrides(&x402.SettlementOverrides{Amount: amount}),
		)
	}
	headers.Del(upstreamSettlementAmountKey)
}

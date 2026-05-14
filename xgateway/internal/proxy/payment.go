package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	mppserver "github.com/tempoxyz/mpp-go/pkg/server"
	chargeserver "github.com/tempoxyz/mpp-go/pkg/tempo/server"
	x402 "github.com/x402-foundation/x402/go"
	x402http "github.com/x402-foundation/x402/go/http"

	"github.com/cp0x-org/xpaywall/xgateway/internal/config"
)

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
		Timeout:   30 * time.Second,
		Transport: &facilitatorLoggingTransport{},
	}
}

const mppRealm = "xgateway"

func runMPPMiddleware(c *gin.Context, configs []mppserver.ComposeConfig) {
	calledNext := false
	var nextRequest *http.Request

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		calledNext = true
		nextRequest = r
	})

	mppserver.ComposeMiddleware(configs...)(next).ServeHTTP(c.Writer, c.Request)
	if !calledNext {
		c.Abort()
		return
	}
	if nextRequest != nil {
		c.Request = nextRequest
	}
	c.Next()
}

func appendMPPChallenges(ctx context.Context, headers http.Header, configs []mppserver.ComposeConfig) {
	for _, cfg := range configs {
		params := cfg.Params
		params.Authorization = ""
		result, err := cfg.Mpp.Charge(ctx, params)
		if err != nil || result == nil || result.Challenge == nil {
			continue
		}
		headers.Add("WWW-Authenticate", result.Challenge.ToAuthenticate(cfg.Mpp.Realm()))
	}
}

func runX402Middleware(c *gin.Context, server *x402http.HTTPServer, decorateHeaders ...func(http.Header)) {
	if server == nil {
		c.Next()
		return
	}

	adapter := &ginAdapter{ctx: c}
	reqCtx := x402http.HTTPRequestContext{
		Adapter: adapter,
		Path:    c.Request.URL.Path,
		Method:  c.Request.Method,
	}

	if !server.RequiresPayment(reqCtx) {
		c.Next()
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result := server.ProcessHTTPRequest(ctx, reqCtx, nil)
	log.Printf("[x402] ProcessHTTPRequest path=%s result=%v", reqCtx.Path, result.Type)
	switch result.Type {
	case x402http.ResultNoPaymentRequired:
		c.Next()
	case x402http.ResultPaymentError:
		if result.Response != nil {
			log.Printf("[x402] payment error: status=%d body=%v", result.Response.Status, result.Response.Body)
		}
		writeX402Error(c, result.Response, decorateHeaders...)
	case x402http.ResultPaymentVerified:
		log.Printf("[x402] payment verified, proceeding to upstream")
		handleX402Verified(c, server, ctx, reqCtx, result)
	}
}

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

func buildMPPServer(cfg config.MPPMethod) (*mppserver.Mpp, error) {
	method, err := chargeserver.MethodFromConfig(chargeserver.Config{
		RPCURL:    cfg.RPCURL,
		Currency:  cfg.Asset,
		Recipient: cfg.Merchant,
	})
	if err != nil {
		return nil, err
	}
	return mppserver.New(method, mppRealm, cfg.SecretKey), nil
}

func parseMPPAmount(price string) string {
	price = strings.TrimSpace(price)
	price = strings.TrimLeftFunc(price, func(r rune) bool {
		return r == '$' || r == '€' || r == '£' || r == '¥'
	})
	return strings.TrimSpace(price)
}

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

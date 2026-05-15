package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tempoxyz/mpp-go/pkg/mpp"
	mppserver "github.com/tempoxyz/mpp-go/pkg/server"
	chargeserver "github.com/tempoxyz/mpp-go/pkg/tempo/server"

	"github.com/cp0x-org/xpaywall/xgateway/internal/config"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

const (
	protoMPP = "mpp"
	mppRealm = "xgateway"
)

// mppProtocol is a PaymentProtocol backed by one or more MPP compose configs.
type mppProtocol struct {
	cfgs []mppserver.ComposeConfig
}

func (p *mppProtocol) Name() string { return protoMPP }

// HasClientAuth reports true when the request carries an MPP payment
// Authorization header — the signal that the client intends to pay via MPP.
func (p *mppProtocol) HasClientAuth(c *gin.Context) bool {
	return mpp.FindPaymentAuthorization(c.GetHeader("Authorization")) != ""
}

// Handle composes the underlying mpp-go middleware. The library writes its own
// 402 response when the payment proof is missing or invalid, so sibling
// protocols cannot decorate that response — decorators are ignored.
func (p *mppProtocol) Handle(c *gin.Context, _ ...func(http.Header)) {
	calledNext := false
	var nextRequest *http.Request

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		calledNext = true
		nextRequest = r
	})

	mppserver.ComposeMiddleware(p.cfgs...)(next).ServeHTTP(c.Writer, c.Request)
	if !calledNext {
		c.Abort()
		return
	}
	if nextRequest != nil {
		c.Request = nextRequest
	}
	c.Next()
}

// Challenge produces a WWW-Authenticate entry for every configured MPP charge,
// allowing a sibling protocol's 402 response to advertise MPP as an option.
func (p *mppProtocol) Challenge(ctx context.Context, headers http.Header) {
	for _, cfg := range p.cfgs {
		params := cfg.Params
		params.Authorization = ""
		result, err := cfg.Mpp.Charge(ctx, params)
		if err != nil || result == nil || result.Challenge == nil {
			continue
		}
		headers.Add("WWW-Authenticate", result.Challenge.ToAuthenticate(cfg.Mpp.Realm()))
	}
}

func newMPPFactory() ProtocolFactory {
	return func(_ context.Context, rule *rules.Rule, _ string, channels []*rules.PaymentChannel) (PaymentProtocol, error) {
		var cfgs []mppserver.ComposeConfig
		for _, ch := range channels {
			price := ch.Price
			if price == "" {
				price = rule.Price
			}
			srv, err := buildMPPServer(config.MPPMethod{
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
			cfgs = append(cfgs, mppserver.ComposeConfig{
				Mpp: srv,
				Params: mppserver.ChargeParams{
					Amount:      parseMPPAmount(price),
					Currency:    ch.ChannelConfig["asset"],
					Recipient:   ch.ChannelConfig["merchant"],
					Description: rule.Description,
				},
			})
		}
		if len(cfgs) == 0 {
			return nil, nil
		}
		return &mppProtocol{cfgs: cfgs}, nil
	}
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

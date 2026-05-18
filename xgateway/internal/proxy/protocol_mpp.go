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

// mppProtocol handles payment verification for one route using the MPP protocol.
type mppProtocol struct {
	cfgs []mppserver.ComposeConfig
}

// HasClientAuth reports whether the request carries an MPP payment Authorization header.
func (p *mppProtocol) HasClientAuth(c *gin.Context) bool {
	return mpp.FindPaymentAuthorization(c.GetHeader("Authorization")) != ""
}

func (p *mppProtocol) Handle(c *gin.Context) {
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

func buildMPPProtocol(_ context.Context, rule *rules.Rule, channels []*rules.PaymentChannel) (*mppProtocol, error) {
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

package gateway

import (
	"encoding/json"
	"math"
	nethttp "net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type Handler struct {
	q *postgres.Queries
}

func New(q *postgres.Queries) *Handler {
	return &Handler{q: q}
}

type resolveRouteResponse struct {
	ProjectID       uuid.UUID       `json:"project_id"`
	OutboundRouteID uuid.UUID       `json:"outbound_route_id"`
	Name            string          `json:"name"`
	InboundPath     string          `json:"inbound_path"`
	Target          string          `json:"target"`
	AuthHeaderName  string          `json:"auth_header_name,omitempty"`
	AuthHeaderValue string          `json:"auth_header_value,omitempty"`
	AllowUnmatched  bool            `json:"allow_unmatched"`
	Price           string          `json:"price"`
	Free            bool            `json:"free"`
	MimeType        string          `json:"mime_type,omitempty"`
	Description     string          `json:"description,omitempty"`
	Bazaar          json.RawMessage `json:"bazaar,omitempty" swaggertype:"object"`
	PaymentChannels []channelDTO    `json:"payment_channels"`
}

type channelDTO struct {
	Protocol        string            `json:"protocol"`
	Method          string            `json:"method,omitempty"`
	Code            string            `json:"code"`
	Scheme          string            `json:"scheme"`
	CaIP2ChainID    string            `json:"caip2_chain_id,omitempty"`
	FacilitatorURL  string            `json:"facilitator_url"`
	PayoutAddress   string            `json:"payout_address,omitempty"`
	AssetSymbol     string            `json:"asset_symbol"`
	ContractAddress string            `json:"contract_address,omitempty"`
	Amount          string            `json:"amount,omitempty"`
	Decimals        int32             `json:"decimals"`
	Enabled         bool              `json:"enabled"`
	PaymentMethodID uuid.UUID         `json:"payment_method_id"`
	AssetID         uuid.UUID         `json:"asset_id"`
	ChannelConfig   map[string]string `json:"channel_config,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ResolveRoute resolves an inbound path to a full proxy rule including target config and payment channels.
// @Summary     Resolve route
// @Tags        gateway
// @Produce     json
// @Param       path path string true "Path in format {projectSlug}/{inboundPath}"
// @Success     200 {object} resolveRouteResponse
// @Failure     404 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    ApiKeyAuth
// @Router      /proxy/resolve/{path} [get]
func (h *Handler) ResolveRoute(c *gin.Context) {
	fullPath := c.Param("path")
	fullPath = strings.TrimPrefix(fullPath, "/")

	parts := strings.SplitN(fullPath, "/", 2)
	if len(parts) < 2 || parts[0] == "" {
		c.JSON(nethttp.StatusNotFound, gin.H{"error": "invalid path: missing project slug"})
		return
	}

	projectSlug := parts[0]
	inboundPath := "/" + parts[1]
	if len(inboundPath) > 1 {
		inboundPath = strings.TrimRight(inboundPath, "/")
	}

	route, err := h.q.ResolveOutboundRoute(c.Request.Context(), postgres.ResolveOutboundRouteParams{
		Slug:        projectSlug,
		InboundPath: inboundPath,
	})
	if err != nil {
		c.JSON(nethttp.StatusNotFound, gin.H{"error": "route not found"})
		return
	}

	priceUSD := route.PriceUsd

	resp := resolveRouteResponse{
		ProjectID:       route.ProjectID,
		OutboundRouteID: route.ID,
		Name:            route.Name,
		InboundPath:     inboundPath,
		Target:          route.BaseUrl,
		AuthHeaderName:  route.AuthHeaderName.String,
		AuthHeaderValue: route.AuthHeaderValue.String,
		AllowUnmatched:  route.AllowUnmatched,
		Price:           priceUSD,
		Free:            route.Free,
		Description:     route.Description,
		Bazaar:          json.RawMessage(route.Bazaar),
		PaymentChannels: []channelDTO{},
	}

	if !route.Free {
		dbMethods, err := h.q.GetPaymentMethodsByProjectSlug(c.Request.Context(), postgres.GetPaymentMethodsByProjectSlugParams{
			Slug:      projectSlug,
			Enabled:   true,
			Enabled_2: true,
		})
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": "failed to load payment methods"})
			return
		}

		for _, m := range dbMethods {
			dto := channelDTO{
				Protocol:        m.Protocol,
				Code:            m.Code,
				Scheme:          m.Scheme,
				FacilitatorURL:  m.FacilitatorUrl.String,
				AssetSymbol:     m.Symbol,
				Decimals:        m.Decimals,
				Amount:          computeRawAmount(priceUSD, m.Decimals),
				Enabled:         m.Enabled,
				PaymentMethodID: m.PaymentMethodID,
				AssetID:         m.AssetID,
			}
			if m.Method.Valid {
				dto.Method = m.Method.String
			}
			if m.Caip2ChainID.Valid {
				dto.CaIP2ChainID = m.Caip2ChainID.String
			}
			if m.PayoutAddress.Valid {
				dto.PayoutAddress = m.PayoutAddress.String
			}
			if m.ContractAddress.Valid {
				dto.ContractAddress = m.ContractAddress.String
			}
			// MPP (Tempo charge) has no facilitator: instead xgateway's HTTP
			// provider builds the charge server from a flat channel_config.
			// rpc_url / secret_key live in the per-project config JSONB; merchant
			// is the payout address and asset is the token contract.
			if m.Protocol == "mpp" {
				dto.ChannelConfig = buildMPPChannelConfig(m.Config, dto.PayoutAddress, dto.ContractAddress)
			}
			resp.PaymentChannels = append(resp.PaymentChannels, dto)
		}
	}

	c.JSON(nethttp.StatusOK, resp)
}

// buildMPPChannelConfig assembles the flat key/value map xgateway's MPP charge
// server expects (rpc_url, secret_key, merchant, asset). rpc_url and secret_key
// are stored in the project_payment_methods.config JSONB; merchant is the payout
// address and asset is the token contract (omitted -> Tempo defaults to the
// chain's stablecoin).
func buildMPPChannelConfig(rawConfig []byte, merchant, asset string) map[string]string {
	cfg := map[string]string{}
	if len(rawConfig) > 0 {
		_ = json.Unmarshal(rawConfig, &cfg)
	}
	if merchant != "" {
		cfg["merchant"] = merchant
	}
	if asset != "" {
		cfg["asset"] = asset
	}
	return cfg
}

// computeRawAmount converts a USD price string (e.g. "0.001") to a raw blockchain
// integer string using the asset's on-chain decimals (e.g. 6 for USDC → "1000").
func computeRawAmount(priceUSD string, decimals int32) string {
	if priceUSD == "" || decimals <= 0 {
		return ""
	}
	f, err := strconv.ParseFloat(priceUSD, 64)
	if err != nil || f <= 0 {
		return ""
	}
	raw := int64(math.Round(f * math.Pow10(int(decimals))))
	return strconv.FormatInt(raw, 10)
}

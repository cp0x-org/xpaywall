package gateway

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
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
	ProjectID       uuid.UUID    `json:"project_id"`
	OutboundRouteID uuid.UUID    `json:"outbound_route_id"`
	Name            string       `json:"name"`
	InboundPath     string       `json:"inbound_path"`
	Target          string       `json:"target"`
	AuthHeaderName  string       `json:"auth_header_name,omitempty"`
	AuthHeaderValue string       `json:"auth_header_value,omitempty"`
	AllowUnmatched  bool         `json:"allow_unmatched"`
	Price           string       `json:"price"`
	Free            bool         `json:"free"`
	MimeType        string       `json:"mime_type,omitempty"`
	Description     string       `json:"description,omitempty"`
	PaymentChannels []channelDTO `json:"payment_channels"`
}

type channelDTO struct {
	Protocol      string            `json:"protocol"`
	Method        string            `json:"method,omitempty"`
	Scheme        string            `json:"scheme"`
	Price         string            `json:"price,omitempty"`
	Enabled       bool              `json:"enabled"`
	ChannelConfig map[string]string `json:"channel_config"`
	ChannelID     uuid.UUID         `json:"channel_id"`
	AssetID       *uuid.UUID        `json:"asset_id,omitempty"`
}

// ResolveRoute resolves an inbound path to a full proxy rule including target config.
// Path format: /{projectSlug}/{inboundPath...}
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

	// Prefer the stored price_usd; fall back to computing from price_amount (cents).
	priceUSD := route.PriceUsd
	if priceUSD == "" && route.PriceAmount > 0 {
		priceUSD = fmt.Sprintf("%.6f", float64(route.PriceAmount)/100.0)
	}

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
		PaymentChannels: []channelDTO{},
	}

	if !route.Free {
		dbChannels, err := h.q.GetPaymentChannelsByProjectSlug(c.Request.Context(), postgres.GetPaymentChannelsByProjectSlugParams{
			Slug:      projectSlug,
			Enabled:   true,
			Enabled_2: true,
		})
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": "failed to load payment channels"})
			return
		}

		for _, ch := range dbChannels {
			cfg := make(map[string]string)
			if len(ch.Metadata) > 0 {
				_ = json.Unmarshal(ch.Metadata, &cfg)
			}
			if ch.PayoutAddress.Valid && ch.PayoutAddress.String != "" {
				cfg["merchant"] = ch.PayoutAddress.String
			}
			dto := channelDTO{
				Protocol:      ch.Protocol,
				Method:        ch.Method,
				Scheme:        ch.Scheme,
				Enabled:       ch.Enabled,
				ChannelConfig: cfg,
				ChannelID:     ch.ChannelID,
			}
			if ch.PaymentChannelAssetID.Valid {
				assetID := uuid.UUID(ch.PaymentChannelAssetID.Bytes)
				dto.AssetID = &assetID
			}
			resp.PaymentChannels = append(resp.PaymentChannels, dto)
		}
	}

	c.JSON(nethttp.StatusOK, resp)
}

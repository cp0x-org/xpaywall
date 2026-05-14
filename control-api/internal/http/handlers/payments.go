package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

// PaymentChannels

type paymentChannelResponse struct {
	ID        uuid.UUID       `json:"id"`
	Protocol  string          `json:"protocol"`
	Method    string          `json:"method"`
	Scheme    string          `json:"scheme"`
	Enabled   bool            `json:"enabled"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func toPaymentChannelResponse(ch postgres.PaymentChannel) paymentChannelResponse {
	var metadata json.RawMessage
	if len(ch.Metadata) > 0 {
		metadata = json.RawMessage(ch.Metadata)
	}
	return paymentChannelResponse{
		ID:        ch.ID,
		Protocol:  ch.Protocol,
		Method:    ch.Method,
		Scheme:    ch.Scheme,
		Enabled:   ch.Enabled,
		Metadata:  metadata,
		CreatedAt: ch.CreatedAt.Time,
		UpdatedAt: ch.UpdatedAt.Time,
	}
}

func (h *Handler) ListPaymentChannels(c *gin.Context) {
	channels, err := h.q.ListPaymentChannels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]paymentChannelResponse, len(channels))
	for i, ch := range channels {
		result[i] = toPaymentChannelResponse(ch)
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetPaymentChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	channel, err := h.q.GetPaymentChannel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment channel not found"})
		return
	}
	c.JSON(http.StatusOK, toPaymentChannelResponse(channel))
}

type createPaymentChannelRequest struct {
	Protocol string          `json:"protocol" binding:"required"`
	Method   string          `json:"method" binding:"required"`
	Scheme   string          `json:"scheme" binding:"required"`
	Enabled  bool            `json:"enabled"`
	Metadata json.RawMessage `json:"metadata"`
}

func (h *Handler) CreatePaymentChannel(c *gin.Context) {
	var req createPaymentChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.q.CreatePaymentChannel(c.Request.Context(), postgres.CreatePaymentChannelParams{
		ID:       uuid.New(),
		Protocol: req.Protocol,
		Method:   req.Method,
		Scheme:   req.Scheme,
		Enabled:  req.Enabled,
		Metadata: []byte(req.Metadata),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toPaymentChannelResponse(channel))
}

type updatePaymentChannelRequest struct {
	Protocol *string         `json:"protocol"`
	Method   *string         `json:"method"`
	Scheme   *string         `json:"scheme"`
	Enabled  *bool           `json:"enabled"`
	Metadata json.RawMessage `json:"metadata"`
}

func (h *Handler) UpdatePaymentChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updatePaymentChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.q.UpdatePaymentChannel(c.Request.Context(), postgres.UpdatePaymentChannelParams{
		ID:       id,
		Protocol: ptrToPgText(req.Protocol),
		Method:   ptrToPgText(req.Method),
		Scheme:   ptrToPgText(req.Scheme),
		Enabled:  boolPtrToPgBool(req.Enabled),
		Metadata: []byte(req.Metadata),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toPaymentChannelResponse(channel))
}

func (h *Handler) DeletePaymentChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeletePaymentChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// PaymentChannelAssets

func (h *Handler) ListPaymentChannelAssets(c *gin.Context) {
	assets, err := h.q.ListPaymentChannelAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assets)
}

func (h *Handler) GetPaymentChannelAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	asset, err := h.q.GetPaymentChannelAsset(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment channel asset not found"})
		return
	}
	c.JSON(http.StatusOK, asset)
}

type createPaymentChannelAssetRequest struct {
	PaymentChannelID uuid.UUID       `json:"payment_channel_id" binding:"required"`
	AssetSymbol      string          `json:"asset_symbol" binding:"required"`
	AssetAddress     *string         `json:"asset_address"`
	Decimals         *int32          `json:"decimals"`
	Metadata         json.RawMessage `json:"metadata"`
}

func (h *Handler) CreatePaymentChannelAsset(c *gin.Context) {
	var req createPaymentChannelAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := postgres.CreatePaymentChannelAssetParams{
		ID:               uuid.New(),
		PaymentChannelID: req.PaymentChannelID,
		AssetSymbol:      req.AssetSymbol,
		Metadata:         []byte(req.Metadata),
	}
	if req.AssetAddress != nil {
		params.AssetAddress = pgtype.Text{String: *req.AssetAddress, Valid: true}
	}
	if req.Decimals != nil {
		params.Decimals = pgtype.Int4{Int32: *req.Decimals, Valid: true}
	}

	asset, err := h.q.CreatePaymentChannelAsset(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, asset)
}

type updatePaymentChannelAssetRequest struct {
	AssetSymbol  *string         `json:"asset_symbol"`
	AssetAddress *string         `json:"asset_address"`
	Decimals     *int32          `json:"decimals"`
	Metadata     json.RawMessage `json:"metadata"`
}

func (h *Handler) UpdatePaymentChannelAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updatePaymentChannelAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset, err := h.q.UpdatePaymentChannelAsset(c.Request.Context(), postgres.UpdatePaymentChannelAssetParams{
		ID:           id,
		AssetSymbol:  ptrToPgText(req.AssetSymbol),
		AssetAddress: ptrToPgText(req.AssetAddress),
		Decimals:     int32PtrToPgInt4(req.Decimals),
		Metadata:     []byte(req.Metadata),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asset)
}

func (h *Handler) DeletePaymentChannelAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeletePaymentChannelAsset(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

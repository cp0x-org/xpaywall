package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

// PaymentChannels

type paymentChannelResponse struct {
	ID        uuid.UUID `json:"id"`
	Protocol  string    `json:"protocol"`
	Method    string    `json:"method"`
	Scheme    string    `json:"scheme"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toPaymentChannelResponse(ch postgres.PaymentChannel) paymentChannelResponse {
	return paymentChannelResponse{
		ID:        ch.ID,
		Protocol:  ch.Protocol,
		Method:    ch.Method,
		Scheme:    ch.Scheme,
		Enabled:   ch.Enabled,
		CreatedAt: ch.CreatedAt.Time,
		UpdatedAt: ch.UpdatedAt.Time,
	}
}

// ListPaymentChannels returns all payment channels.
// @Summary     List payment channels
// @Tags        payment-channels
// @Produce     json
// @Success     200 {array} paymentChannelResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channels [get]
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

// GetPaymentChannel returns a payment channel by ID.
// @Summary     Get payment channel
// @Tags        payment-channels
// @Produce     json
// @Param       id path string true "Payment Channel ID (UUID)"
// @Success     200 {object} paymentChannelResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channels/{id} [get]
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
	Protocol string `json:"protocol" binding:"required"`
	Method   string `json:"method" binding:"required"`
	Scheme   string `json:"scheme" binding:"required"`
	Enabled  bool   `json:"enabled"`
}

// CreatePaymentChannel creates a new payment channel.
// @Summary     Create payment channel
// @Tags        payment-channels
// @Accept      json
// @Produce     json
// @Param       body body createPaymentChannelRequest true "Payment channel data"
// @Success     201 {object} paymentChannelResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channels [post]
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
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toPaymentChannelResponse(channel))
}

type updatePaymentChannelRequest struct {
	Protocol *string `json:"protocol"`
	Method   *string `json:"method"`
	Scheme   *string `json:"scheme"`
	Enabled  *bool   `json:"enabled"`
}

// UpdatePaymentChannel updates a payment channel by ID.
// @Summary     Update payment channel
// @Tags        payment-channels
// @Accept      json
// @Produce     json
// @Param       id path string true "Payment Channel ID (UUID)"
// @Param       body body updatePaymentChannelRequest true "Fields to update"
// @Success     200 {object} paymentChannelResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channels/{id} [put]
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
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toPaymentChannelResponse(channel))
}

// DeletePaymentChannel deletes a payment channel by ID.
// @Summary     Delete payment channel
// @Tags        payment-channels
// @Param       id path string true "Payment Channel ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channels/{id} [delete]
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

// ListPaymentChannelAssets returns all payment channel assets.
// @Summary     List payment channel assets
// @Tags        payment-channel-assets
// @Produce     json
// @Success     200 {array} object "Array of payment channel asset objects"
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channel-assets [get]
func (h *Handler) ListPaymentChannelAssets(c *gin.Context) {
	assets, err := h.q.ListPaymentChannelAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// GetPaymentChannelAsset returns a payment channel asset by ID.
// @Summary     Get payment channel asset
// @Tags        payment-channel-assets
// @Produce     json
// @Param       id path string true "Asset ID (UUID)"
// @Success     200 {object} object "Payment channel asset object"
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channel-assets/{id} [get]
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
	PaymentChannelID uuid.UUID `json:"payment_channel_id" binding:"required"`
	AssetSymbol      string    `json:"asset_symbol" binding:"required"`
	AssetAddress     *string   `json:"asset_address"`
	Decimals         *int32    `json:"decimals"`
}

// CreatePaymentChannelAsset creates a new payment channel asset.
// @Summary     Create payment channel asset
// @Tags        payment-channel-assets
// @Accept      json
// @Produce     json
// @Param       body body createPaymentChannelAssetRequest true "Asset data"
// @Success     201 {object} object "Created asset object"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channel-assets [post]
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
	AssetSymbol  *string `json:"asset_symbol"`
	AssetAddress *string `json:"asset_address"`
	Decimals     *int32  `json:"decimals"`
}

// UpdatePaymentChannelAsset updates a payment channel asset by ID.
// @Summary     Update payment channel asset
// @Tags        payment-channel-assets
// @Accept      json
// @Produce     json
// @Param       id path string true "Asset ID (UUID)"
// @Param       body body updatePaymentChannelAssetRequest true "Fields to update"
// @Success     200 {object} object "Updated asset object"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channel-assets/{id} [put]
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
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asset)
}

// DeletePaymentChannelAsset deletes a payment channel asset by ID.
// @Summary     Delete payment channel asset
// @Tags        payment-channel-assets
// @Param       id path string true "Asset ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-channel-assets/{id} [delete]
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

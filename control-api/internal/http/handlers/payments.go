package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

// ─── Payment Methods ──────────────────────────────────────────────────────────

type paymentMethodResponse struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Protocol     string    `json:"protocol"`
	Name         string    `json:"name"`
	Caip2ChainID *string   `json:"caip2_chain_id,omitempty"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toPaymentMethodResponse(m postgres.PaymentMethod) paymentMethodResponse {
	return paymentMethodResponse{
		ID:           m.ID,
		Code:         m.Code,
		Protocol:     m.Protocol,
		Name:         m.Name,
		Caip2ChainID: pgTextPtr(m.Caip2ChainID),
		Enabled:      m.Enabled,
		CreatedAt:    m.CreatedAt.Time,
		UpdatedAt:    m.UpdatedAt.Time,
	}
}

// ListPaymentMethods returns all payment methods.
// @Summary     List payment methods
// @Tags        payment-methods
// @Produce     json
// @Success     200 {array} paymentMethodResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-methods [get]
func (h *Handler) ListPaymentMethods(c *gin.Context) {
	methods, err := h.q.ListPaymentMethods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]paymentMethodResponse, len(methods))
	for i, m := range methods {
		result[i] = toPaymentMethodResponse(m)
	}
	c.JSON(http.StatusOK, result)
}

// GetPaymentMethod returns a payment method by ID.
// @Summary     Get payment method
// @Tags        payment-methods
// @Produce     json
// @Param       id path string true "Payment Method ID (UUID)"
// @Success     200 {object} paymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-methods/{id} [get]
func (h *Handler) GetPaymentMethod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, err := h.q.GetPaymentMethod(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment method not found"})
		return
	}
	c.JSON(http.StatusOK, toPaymentMethodResponse(m))
}

type createPaymentMethodRequest struct {
	Code         string  `json:"code" binding:"required"`
	Protocol     string  `json:"protocol" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Caip2ChainID *string `json:"caip2_chain_id"`
	Enabled      bool    `json:"enabled"`
}

// CreatePaymentMethod creates a new payment method.
// @Summary     Create payment method
// @Tags        payment-methods
// @Accept      json
// @Produce     json
// @Param       body body createPaymentMethodRequest true "Payment method data"
// @Success     201 {object} paymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-methods [post]
func (h *Handler) CreatePaymentMethod(c *gin.Context) {
	var req createPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.q.CreatePaymentMethod(c.Request.Context(), postgres.CreatePaymentMethodParams{
		ID:           uuid.New(),
		Code:         req.Code,
		Protocol:     req.Protocol,
		Name:         req.Name,
		Caip2ChainID: ptrToPgText(req.Caip2ChainID),
		Enabled:      req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toPaymentMethodResponse(m))
}

type updatePaymentMethodRequest struct {
	Code         *string `json:"code"`
	Protocol     *string `json:"protocol"`
	Name         *string `json:"name"`
	Caip2ChainID *string `json:"caip2_chain_id"`
	Enabled      *bool   `json:"enabled"`
}

// UpdatePaymentMethod updates a payment method by ID.
// @Summary     Update payment method
// @Tags        payment-methods
// @Accept      json
// @Produce     json
// @Param       id path string true "Payment Method ID (UUID)"
// @Param       body body updatePaymentMethodRequest true "Fields to update"
// @Success     200 {object} paymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-methods/{id} [put]
func (h *Handler) UpdatePaymentMethod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.q.UpdatePaymentMethod(c.Request.Context(), postgres.UpdatePaymentMethodParams{
		ID:           id,
		Code:         ptrToPgText(req.Code),
		Protocol:     ptrToPgText(req.Protocol),
		Name:         ptrToPgText(req.Name),
		Caip2ChainID: ptrToPgText(req.Caip2ChainID),
		Enabled:      boolPtrToPgBool(req.Enabled),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toPaymentMethodResponse(m))
}

// DeletePaymentMethod deletes a payment method by ID.
// @Summary     Delete payment method
// @Tags        payment-methods
// @Param       id path string true "Payment Method ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-methods/{id} [delete]
func (h *Handler) DeletePaymentMethod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeletePaymentMethod(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Payment Method Assets ────────────────────────────────────────────────────

type paymentMethodAssetResponse struct {
	ID                 uuid.UUID `json:"id"`
	PaymentMethodID    uuid.UUID `json:"payment_method_id"`
	PaymentMethodName  string    `json:"payment_method_name"`
	PaymentMethodChain *string   `json:"payment_method_chain,omitempty"`
	Symbol             string    `json:"symbol"`
	ContractAddress    *string   `json:"contract_address,omitempty"`
	Decimals           int32     `json:"decimals"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func toPaymentMethodAssetResponse(a postgres.ListPaymentMethodAssetsRow) paymentMethodAssetResponse {
	return paymentMethodAssetResponse{
		ID:                 a.ID,
		PaymentMethodID:    a.PaymentMethodID,
		PaymentMethodName:  a.PaymentMethodName,
		PaymentMethodChain: pgTextPtr(a.PaymentMethodChain),
		Symbol:             a.Symbol,
		ContractAddress:    pgTextPtr(a.ContractAddress),
		Decimals:           a.Decimals,
		CreatedAt:          a.CreatedAt.Time,
		UpdatedAt:          a.UpdatedAt.Time,
	}
}

func toPaymentMethodAssetResponseFromGet(a postgres.GetPaymentMethodAssetRow) paymentMethodAssetResponse {
	return paymentMethodAssetResponse{
		ID:                 a.ID,
		PaymentMethodID:    a.PaymentMethodID,
		PaymentMethodName:  a.PaymentMethodName,
		PaymentMethodChain: pgTextPtr(a.PaymentMethodChain),
		Symbol:             a.Symbol,
		ContractAddress:    pgTextPtr(a.ContractAddress),
		Decimals:           a.Decimals,
		CreatedAt:          a.CreatedAt.Time,
		UpdatedAt:          a.UpdatedAt.Time,
	}
}

// ListPaymentMethodAssets returns all payment method assets.
// @Summary     List payment method assets
// @Tags        payment-method-assets
// @Produce     json
// @Success     200 {array} paymentMethodAssetResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-method-assets [get]
func (h *Handler) ListPaymentMethodAssets(c *gin.Context) {
	assets, err := h.q.ListPaymentMethodAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]paymentMethodAssetResponse, len(assets))
	for i, a := range assets {
		result[i] = toPaymentMethodAssetResponse(a)
	}
	c.JSON(http.StatusOK, result)
}

// GetPaymentMethodAsset returns a payment method asset by ID.
// @Summary     Get payment method asset
// @Tags        payment-method-assets
// @Produce     json
// @Param       id path string true "Asset ID (UUID)"
// @Success     200 {object} paymentMethodAssetResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-method-assets/{id} [get]
func (h *Handler) GetPaymentMethodAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	a, err := h.q.GetPaymentMethodAsset(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment method asset not found"})
		return
	}
	c.JSON(http.StatusOK, toPaymentMethodAssetResponseFromGet(a))
}

type createPaymentMethodAssetRequest struct {
	PaymentMethodID uuid.UUID `json:"payment_method_id" binding:"required"`
	Symbol          string    `json:"symbol" binding:"required"`
	ContractAddress *string   `json:"contract_address"`
	Decimals        int32     `json:"decimals" binding:"required"`
}

// CreatePaymentMethodAsset creates a new payment method asset.
// @Summary     Create payment method asset
// @Tags        payment-method-assets
// @Accept      json
// @Produce     json
// @Param       body body createPaymentMethodAssetRequest true "Asset data"
// @Success     201 {object} paymentMethodAssetResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-method-assets [post]
func (h *Handler) CreatePaymentMethodAsset(c *gin.Context) {
	var req createPaymentMethodAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.q.CreatePaymentMethodAsset(c.Request.Context(), postgres.CreatePaymentMethodAssetParams{
		ID:              uuid.New(),
		PaymentMethodID: req.PaymentMethodID,
		Symbol:          req.Symbol,
		ContractAddress: ptrToPgText(req.ContractAddress),
		Decimals:        req.Decimals,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a, err := h.q.GetPaymentMethodAsset(c.Request.Context(), created.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toPaymentMethodAssetResponseFromGet(a))
}

type updatePaymentMethodAssetRequest struct {
	Symbol          *string `json:"symbol"`
	ContractAddress *string `json:"contract_address"`
	Decimals        *int32  `json:"decimals"`
}

// UpdatePaymentMethodAsset updates a payment method asset by ID.
// @Summary     Update payment method asset
// @Tags        payment-method-assets
// @Accept      json
// @Produce     json
// @Param       id path string true "Asset ID (UUID)"
// @Param       body body updatePaymentMethodAssetRequest true "Fields to update"
// @Success     200 {object} paymentMethodAssetResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-method-assets/{id} [put]
func (h *Handler) UpdatePaymentMethodAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updatePaymentMethodAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = h.q.UpdatePaymentMethodAsset(c.Request.Context(), postgres.UpdatePaymentMethodAssetParams{
		ID:              id,
		Symbol:          ptrToPgText(req.Symbol),
		ContractAddress: ptrToPgText(req.ContractAddress),
		Decimals:        int32PtrToPgInt4(req.Decimals),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a, err := h.q.GetPaymentMethodAsset(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toPaymentMethodAssetResponseFromGet(a))
}

// DeletePaymentMethodAsset deletes a payment method asset by ID.
// @Summary     Delete payment method asset
// @Tags        payment-method-assets
// @Param       id path string true "Asset ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/payment-method-assets/{id} [delete]
func (h *Handler) DeletePaymentMethodAsset(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeletePaymentMethodAsset(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Facilitators ─────────────────────────────────────────────────────────────

type facilitatorResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toFacilitatorResponse(f postgres.Facilitator) facilitatorResponse {
	return facilitatorResponse{
		ID:        f.ID,
		Name:      f.Name,
		URL:       f.Url,
		Enabled:   f.Enabled,
		CreatedAt: f.CreatedAt.Time,
		UpdatedAt: f.UpdatedAt.Time,
	}
}

// ListFacilitators returns all facilitators.
// @Summary     List facilitators
// @Tags        facilitators
// @Produce     json
// @Success     200 {array} facilitatorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/facilitators [get]
func (h *Handler) ListFacilitators(c *gin.Context) {
	facilitators, err := h.q.ListFacilitators(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]facilitatorResponse, len(facilitators))
	for i, f := range facilitators {
		result[i] = toFacilitatorResponse(f)
	}
	c.JSON(http.StatusOK, result)
}

// GetFacilitator returns a facilitator by ID.
// @Summary     Get facilitator
// @Tags        facilitators
// @Produce     json
// @Param       id path string true "Facilitator ID (UUID)"
// @Success     200 {object} facilitatorResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/facilitators/{id} [get]
func (h *Handler) GetFacilitator(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	f, err := h.q.GetFacilitator(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "facilitator not found"})
		return
	}
	c.JSON(http.StatusOK, toFacilitatorResponse(f))
}

type createFacilitatorRequest struct {
	Name    string `json:"name" binding:"required"`
	URL     string `json:"url" binding:"required"`
	Enabled bool   `json:"enabled"`
}

// CreateFacilitator creates a new facilitator.
// @Summary     Create facilitator
// @Tags        facilitators
// @Accept      json
// @Produce     json
// @Param       body body createFacilitatorRequest true "Facilitator data"
// @Success     201 {object} facilitatorResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/facilitators [post]
func (h *Handler) CreateFacilitator(c *gin.Context) {
	var req createFacilitatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := h.q.CreateFacilitator(c.Request.Context(), postgres.CreateFacilitatorParams{
		ID:      uuid.New(),
		Name:    req.Name,
		Url:     req.URL,
		Enabled: req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toFacilitatorResponse(f))
}

type updateFacilitatorRequest struct {
	Name    *string `json:"name"`
	URL     *string `json:"url"`
	Enabled *bool   `json:"enabled"`
}

// UpdateFacilitator updates a facilitator by ID.
// @Summary     Update facilitator
// @Tags        facilitators
// @Accept      json
// @Produce     json
// @Param       id path string true "Facilitator ID (UUID)"
// @Param       body body updateFacilitatorRequest true "Fields to update"
// @Success     200 {object} facilitatorResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/facilitators/{id} [put]
func (h *Handler) UpdateFacilitator(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateFacilitatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := h.q.UpdateFacilitator(c.Request.Context(), postgres.UpdateFacilitatorParams{
		ID:      id,
		Name:    ptrToPgText(req.Name),
		Url:     ptrToPgText(req.URL),
		Enabled: boolPtrToPgBool(req.Enabled),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toFacilitatorResponse(f))
}

// DeleteFacilitator deletes a facilitator by ID.
// @Summary     Delete facilitator
// @Tags        facilitators
// @Param       id path string true "Facilitator ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/facilitators/{id} [delete]
func (h *Handler) DeleteFacilitator(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeleteFacilitator(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Project Payment Methods ──────────────────────────────────────────────────

type projectPaymentMethodResponse struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"project_id"`
	PaymentMethodID uuid.UUID `json:"payment_method_id"`
	AssetID         uuid.UUID `json:"asset_id"`
	Scheme          string    `json:"scheme"`
	FacilitatorID   uuid.UUID `json:"facilitator_id"`
	PayoutAddress   *string   `json:"payout_address,omitempty"`
	Config          []byte    `json:"config,omitempty"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toProjectPaymentMethodResponse(p postgres.ProjectPaymentMethod) projectPaymentMethodResponse {
	return projectPaymentMethodResponse{
		ID:              p.ID,
		ProjectID:       p.ProjectID,
		PaymentMethodID: p.PaymentMethodID,
		AssetID:         p.AssetID,
		Scheme:          p.Scheme,
		FacilitatorID:   p.FacilitatorID,
		PayoutAddress:   pgTextPtr(p.PayoutAddress),
		Config:          p.Config,
		Enabled:         p.Enabled,
		CreatedAt:       p.CreatedAt.Time,
		UpdatedAt:       p.UpdatedAt.Time,
	}
}

// ListProjectPaymentMethods lists payment methods for a project.
// @Summary     List project payment methods
// @Tags        project-payment-methods
// @Produce     json
// @Param       project_id query string true "Project ID (UUID)"
// @Success     200 {array} projectPaymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/project-payment-methods [get]
func (h *Handler) ListProjectPaymentMethods(c *gin.Context) {
	pidStr := c.Query("project_id")
	pid, err := uuid.Parse(pidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	rows, err := h.q.ListProjectPaymentMethods(c.Request.Context(), pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]projectPaymentMethodResponse, len(rows))
	for i, r := range rows {
		result[i] = toProjectPaymentMethodResponse(r)
	}
	c.JSON(http.StatusOK, result)
}

type projectPaymentMethodFullResponse struct {
	ID                uuid.UUID `json:"id"`
	ProjectID         uuid.UUID `json:"project_id"`
	ProjectName       string    `json:"project_name"`
	PaymentMethodID   uuid.UUID `json:"payment_method_id"`
	PaymentMethodName string    `json:"payment_method_name"`
	AssetID           uuid.UUID `json:"asset_id"`
	AssetSymbol       string    `json:"asset_symbol"`
	Scheme            string    `json:"scheme"`
	FacilitatorID     uuid.UUID `json:"facilitator_id"`
	FacilitatorName   string    `json:"facilitator_name"`
	PayoutAddress     *string   `json:"payout_address,omitempty"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toProjectPaymentMethodFullResponse(r postgres.ListAllProjectPaymentMethodsRow) projectPaymentMethodFullResponse {
	return projectPaymentMethodFullResponse{
		ID:                r.ID,
		ProjectID:         r.ProjectID,
		ProjectName:       r.ProjectName,
		PaymentMethodID:   r.PaymentMethodID,
		PaymentMethodName: r.PaymentMethodName,
		AssetID:           r.AssetID,
		AssetSymbol:       r.AssetSymbol,
		Scheme:            r.Scheme,
		FacilitatorID:     r.FacilitatorID,
		FacilitatorName:   r.FacilitatorName,
		PayoutAddress:     pgTextPtr(r.PayoutAddress),
		Enabled:           r.Enabled,
		CreatedAt:         r.CreatedAt.Time,
		UpdatedAt:         r.UpdatedAt.Time,
	}
}

// ListAllProjectPaymentMethods lists all project payment methods across all projects with joined names.
// @Summary     List all project payment methods
// @Tags        project-payment-methods
// @Produce     json
// @Success     200 {array} projectPaymentMethodFullResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/project-payment-methods/all [get]
func (h *Handler) ListAllProjectPaymentMethods(c *gin.Context) {
	rows, err := h.q.ListAllProjectPaymentMethods(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]projectPaymentMethodFullResponse, len(rows))
	for i, r := range rows {
		result[i] = toProjectPaymentMethodFullResponse(r)
	}
	c.JSON(http.StatusOK, result)
}

// GetProjectPaymentMethod returns a project payment method by ID.
// @Summary     Get project payment method
// @Tags        project-payment-methods
// @Produce     json
// @Param       id path string true "Project Payment Method ID (UUID)"
// @Success     200 {object} projectPaymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/project-payment-methods/{id} [get]
func (h *Handler) GetProjectPaymentMethod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row, err := h.q.GetProjectPaymentMethod(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project payment method not found"})
		return
	}
	c.JSON(http.StatusOK, toProjectPaymentMethodResponse(row))
}

type createProjectPaymentMethodRequest struct {
	ProjectID       uuid.UUID `json:"project_id" binding:"required"`
	PaymentMethodID uuid.UUID `json:"payment_method_id" binding:"required"`
	AssetID         uuid.UUID `json:"asset_id" binding:"required"`
	Scheme          string    `json:"scheme" binding:"required"`
	FacilitatorID   uuid.UUID `json:"facilitator_id" binding:"required"`
	PayoutAddress   *string   `json:"payout_address"`
	Config          []byte    `json:"config"`
	Enabled         bool      `json:"enabled"`
}

// CreateProjectPaymentMethod creates a project payment method.
// @Summary     Create project payment method
// @Tags        project-payment-methods
// @Accept      json
// @Produce     json
// @Param       body body createProjectPaymentMethodRequest true "Project payment method data"
// @Success     201 {object} projectPaymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/project-payment-methods [post]
func (h *Handler) CreateProjectPaymentMethod(c *gin.Context) {
	var req createProjectPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.q.CreateProjectPaymentMethod(c.Request.Context(), postgres.CreateProjectPaymentMethodParams{
		ID:              uuid.New(),
		ProjectID:       req.ProjectID,
		PaymentMethodID: req.PaymentMethodID,
		AssetID:         req.AssetID,
		Scheme:          req.Scheme,
		FacilitatorID:   req.FacilitatorID,
		PayoutAddress:   ptrToPgText(req.PayoutAddress),
		Config:          req.Config,
		Enabled:         req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toProjectPaymentMethodResponse(row))
}

type updateProjectPaymentMethodRequest struct {
	Scheme        *string    `json:"scheme"`
	FacilitatorID *uuid.UUID `json:"facilitator_id"`
	PayoutAddress *string    `json:"payout_address"`
	Config        []byte     `json:"config"`
	Enabled       *bool      `json:"enabled"`
}

// UpdateProjectPaymentMethod updates a project payment method by ID.
// @Summary     Update project payment method
// @Tags        project-payment-methods
// @Accept      json
// @Produce     json
// @Param       id path string true "Project Payment Method ID (UUID)"
// @Param       body body updateProjectPaymentMethodRequest true "Fields to update"
// @Success     200 {object} projectPaymentMethodResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/project-payment-methods/{id} [put]
func (h *Handler) UpdateProjectPaymentMethod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateProjectPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.q.UpdateProjectPaymentMethod(c.Request.Context(), postgres.UpdateProjectPaymentMethodParams{
		ID:            id,
		Scheme:        ptrToPgText(req.Scheme),
		FacilitatorID: req.FacilitatorID,
		PayoutAddress: ptrToPgText(req.PayoutAddress),
		Config:        req.Config,
		Enabled:       boolPtrToPgBool(req.Enabled),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toProjectPaymentMethodResponse(row))
}

// DeleteProjectPaymentMethod deletes a project payment method by ID.
// @Summary     Delete project payment method
// @Tags        project-payment-methods
// @Param       id path string true "Project Payment Method ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/project-payment-methods/{id} [delete]
func (h *Handler) DeleteProjectPaymentMethod(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeleteProjectPaymentMethod(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

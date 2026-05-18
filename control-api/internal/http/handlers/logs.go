package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

func pgUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func uuidPtrToPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(*u), Valid: true}
}

func pgInt4ToPtr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

func int32PtrToPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func pgInt8ToPtr(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

func int64PtrToPgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

func pgTimestampToPtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func timePtrToPgTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

func pgNumericToStringPtr(n pgtype.Numeric) *string {
	if !n.Valid {
		return nil
	}
	f8, err := n.Float64Value()
	if err != nil || !f8.Valid {
		return nil
	}
	s := strconv.FormatFloat(f8.Float64, 'f', 8, 64)
	return &s
}

func stringPtrToPgNumeric(s *string) pgtype.Numeric {
	var n pgtype.Numeric
	if s == nil {
		return n
	}
	v := strings.TrimLeft(*s, "$")
	_ = n.Scan(v)
	return n
}

// ─── Response types ───────────────────────────────────────────────────────────

type requestLogResponse struct {
	ID                     uuid.UUID  `json:"id"`
	ProjectID              uuid.UUID  `json:"project_id"`
	OutboundRouteID        *uuid.UUID `json:"outbound_route_id,omitempty"`
	RequestID              string     `json:"request_id"`
	Method                 string     `json:"method"`
	Path                   string     `json:"path"`
	ClientIP               *string    `json:"client_ip,omitempty"`
	UserAgent              *string    `json:"user_agent,omitempty"`
	Status                 string     `json:"status"`
	PaymentRequired        bool       `json:"payment_required"`
	PaymentRequestedAt     *time.Time `json:"payment_requested_at,omitempty"`
	PaymentCompleted       bool       `json:"payment_completed"`
	PaymentCompletedAt     *time.Time `json:"payment_completed_at,omitempty"`
	PaymentChannelID       *uuid.UUID `json:"payment_channel_id,omitempty"`
	PaymentChannelAssetID  *uuid.UUID `json:"payment_channel_asset_id,omitempty"`
	AmountPaid             *int64     `json:"amount_paid,omitempty"`
	AmountUSD              *string    `json:"amount_usd,omitempty"`
	UpstreamURL            *string    `json:"upstream_url,omitempty"`
	UpstreamStatusCode     *int32     `json:"upstream_status_code,omitempty"`
	UpstreamResponseTimeMs *int32     `json:"upstream_response_time_ms,omitempty"`
	FinalStatusCode        *int32     `json:"final_status_code,omitempty"`
	ErrorType              *string    `json:"error_type,omitempty"`
	ErrorMessage           *string    `json:"error_message,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type requestEventResponse struct {
	ID           uuid.UUID       `json:"id"`
	RequestLogID uuid.UUID       `json:"request_log_id"`
	EventType    string          `json:"event_type"`
	Metadata     json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt    time.Time       `json:"created_at"`
}

func toRequestLogResponse(r postgres.RequestLog) requestLogResponse {
	return requestLogResponse{
		ID:                     r.ID,
		ProjectID:              r.ProjectID,
		OutboundRouteID:        pgUUIDToPtr(r.OutboundRouteID),
		RequestID:              r.RequestID,
		Method:                 r.Method,
		Path:                   r.Path,
		ClientIP:               pgTextPtr(r.ClientIp),
		UserAgent:              r.UserAgent,
		Status:                 r.Status,
		PaymentRequired:        r.PaymentRequired,
		PaymentRequestedAt:     pgTimestampToPtr(r.PaymentRequestedAt),
		PaymentCompleted:       r.PaymentCompleted,
		PaymentCompletedAt:     pgTimestampToPtr(r.PaymentCompletedAt),
		PaymentChannelID:       pgUUIDToPtr(r.PaymentChannelID),
		PaymentChannelAssetID:  pgUUIDToPtr(r.PaymentChannelAssetID),
		AmountPaid:             pgInt8ToPtr(r.AmountPaid),
		AmountUSD:              pgNumericToStringPtr(r.AmountUsd),
		UpstreamURL:            r.UpstreamUrl,
		UpstreamStatusCode:     pgInt4ToPtr(r.UpstreamStatusCode),
		UpstreamResponseTimeMs: pgInt4ToPtr(r.UpstreamResponseTimeMs),
		FinalStatusCode:        pgInt4ToPtr(r.FinalStatusCode),
		ErrorType:              pgTextPtr(r.ErrorType),
		ErrorMessage:           r.ErrorMessage,
		CreatedAt:              r.CreatedAt.Time,
		UpdatedAt:              r.UpdatedAt.Time,
	}
}

func toRequestEventResponse(e postgres.RequestEvent) requestEventResponse {
	return requestEventResponse{
		ID:           e.ID,
		RequestLogID: e.RequestLogID,
		EventType:    e.EventType,
		Metadata:     json.RawMessage(e.Metadata),
		CreatedAt:    e.CreatedAt.Time,
	}
}

// ─── Request Logs ─────────────────────────────────────────────────────────────

// ListRequestLogs returns request logs with optional pagination and project filter.
// @Summary     List request logs
// @Tags        request-logs
// @Produce     json
// @Param       limit query int false "Max results (default 50)"
// @Param       offset query int false "Offset for pagination (default 0)"
// @Param       project_id query string false "Filter by Project ID (UUID)"
// @Success     200 {array} requestLogResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/request-logs [get]
func (h *Handler) ListRequestLogs(c *gin.Context) {
	limit := int32(50)
	offset := int32(0)
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.ParseInt(l, 10, 32); err == nil {
			limit = int32(v)
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.ParseInt(o, 10, 32); err == nil {
			offset = int32(v)
		}
	}

	if pid := c.Query("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		rows, err := h.q.ListRequestLogsByProject(c.Request.Context(), postgres.ListRequestLogsByProjectParams{
			ProjectID: id,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result := make([]requestLogResponse, len(rows))
		for i, r := range rows {
			result[i] = toRequestLogResponse(r)
		}
		c.JSON(http.StatusOK, result)
		return
	}

	rows, err := h.q.ListRequestLogs(c.Request.Context(), postgres.ListRequestLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]requestLogResponse, len(rows))
	for i, r := range rows {
		result[i] = toRequestLogResponse(r)
	}
	c.JSON(http.StatusOK, result)
}

// GetRequestLog returns a single request log by ID.
// @Summary     Get request log
// @Tags        request-logs
// @Produce     json
// @Param       id path string true "Request Log ID (UUID)"
// @Success     200 {object} requestLogResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/request-logs/{id} [get]
func (h *Handler) GetRequestLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row, err := h.q.GetRequestLog(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, toRequestLogResponse(row))
}

type createRequestLogRequest struct {
	ID                     *uuid.UUID `json:"id"`
	ProjectID              uuid.UUID  `json:"project_id" binding:"required"`
	OutboundRouteID        *uuid.UUID `json:"outbound_route_id"`
	RequestID              string     `json:"request_id" binding:"required"`
	Method                 string     `json:"method" binding:"required"`
	Path                   string     `json:"path" binding:"required"`
	ClientIP               *string    `json:"client_ip"`
	UserAgent              *string    `json:"user_agent"`
	Status                 string     `json:"status" binding:"required"`
	PaymentRequired        bool       `json:"payment_required"`
	PaymentRequestedAt     *time.Time `json:"payment_requested_at"`
	PaymentCompleted       bool       `json:"payment_completed"`
	PaymentCompletedAt     *time.Time `json:"payment_completed_at"`
	PaymentChannelID       *uuid.UUID `json:"payment_channel_id"`
	PaymentChannelAssetID  *uuid.UUID `json:"payment_channel_asset_id"`
	AmountPaid             *int64     `json:"amount_paid"`
	AmountUSD              *string    `json:"amount_usd"`
	UpstreamURL            *string    `json:"upstream_url"`
	UpstreamStatusCode     *int32     `json:"upstream_status_code"`
	UpstreamResponseTimeMs *int32     `json:"upstream_response_time_ms"`
	FinalStatusCode        *int32     `json:"final_status_code"`
	ErrorType              *string    `json:"error_type"`
	ErrorMessage           *string    `json:"error_message"`
}

// CreateRequestLog creates a new request log entry (internal — called by xgateway).
// @Summary     Create request log
// @Tags        request-logs
// @Accept      json
// @Produce     json
// @Param       body body createRequestLogRequest true "Request log data"
// @Success     201 {object} requestLogResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    ApiKeyAuth
// @Router      /api/v1/request-logs [post]
func (h *Handler) CreateRequestLog(c *gin.Context) {
	var req createRequestLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := uuid.New()
	if req.ID != nil {
		id = *req.ID
	}
	row, err := h.q.CreateRequestLog(c.Request.Context(), postgres.CreateRequestLogParams{
		ID:                     id,
		ProjectID:              req.ProjectID,
		OutboundRouteID:        uuidPtrToPgUUID(req.OutboundRouteID),
		RequestID:              req.RequestID,
		Method:                 req.Method,
		Path:                   req.Path,
		ClientIp:               ptrToPgText(req.ClientIP),
		UserAgent:              req.UserAgent,
		Status:                 req.Status,
		PaymentRequired:        req.PaymentRequired,
		PaymentRequestedAt:     timePtrToPgTimestamp(req.PaymentRequestedAt),
		PaymentCompleted:       req.PaymentCompleted,
		PaymentCompletedAt:     timePtrToPgTimestamp(req.PaymentCompletedAt),
		PaymentChannelID:       uuidPtrToPgUUID(req.PaymentChannelID),
		PaymentChannelAssetID:  uuidPtrToPgUUID(req.PaymentChannelAssetID),
		AmountPaid:             int64PtrToPgInt8(req.AmountPaid),
		AmountUsd:              stringPtrToPgNumeric(req.AmountUSD),
		UpstreamUrl:            req.UpstreamURL,
		UpstreamStatusCode:     int32PtrToPgInt4(req.UpstreamStatusCode),
		UpstreamResponseTimeMs: int32PtrToPgInt4(req.UpstreamResponseTimeMs),
		FinalStatusCode:        int32PtrToPgInt4(req.FinalStatusCode),
		ErrorType:              ptrToPgText(req.ErrorType),
		ErrorMessage:           req.ErrorMessage,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toRequestLogResponse(row))
}

type updateRequestLogRequest struct {
	Status                 string     `json:"status" binding:"required"`
	OutboundRouteID        *uuid.UUID `json:"outbound_route_id"`
	PaymentRequired        bool       `json:"payment_required"`
	PaymentRequestedAt     *time.Time `json:"payment_requested_at"`
	PaymentCompleted       bool       `json:"payment_completed"`
	PaymentCompletedAt     *time.Time `json:"payment_completed_at"`
	PaymentChannelID       *uuid.UUID `json:"payment_channel_id"`
	PaymentChannelAssetID  *uuid.UUID `json:"payment_channel_asset_id"`
	AmountPaid             *int64     `json:"amount_paid"`
	AmountUSD              *string    `json:"amount_usd"`
	UpstreamURL            *string    `json:"upstream_url"`
	UpstreamStatusCode     *int32     `json:"upstream_status_code"`
	UpstreamResponseTimeMs *int32     `json:"upstream_response_time_ms"`
	FinalStatusCode        *int32     `json:"final_status_code"`
	ErrorType              *string    `json:"error_type"`
	ErrorMessage           *string    `json:"error_message"`
}

// UpdateRequestLog updates a request log entry (internal — called by xgateway).
// @Summary     Update request log
// @Tags        request-logs
// @Accept      json
// @Produce     json
// @Param       id path string true "Request Log ID (UUID)"
// @Param       body body updateRequestLogRequest true "Fields to update"
// @Success     200 {object} requestLogResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    ApiKeyAuth
// @Router      /api/v1/request-logs/{id} [put]
func (h *Handler) UpdateRequestLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateRequestLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.q.UpdateRequestLog(c.Request.Context(), postgres.UpdateRequestLogParams{
		ID:                     id,
		Status:                 req.Status,
		OutboundRouteID:        uuidPtrToPgUUID(req.OutboundRouteID),
		PaymentRequired:        req.PaymentRequired,
		PaymentRequestedAt:     timePtrToPgTimestamp(req.PaymentRequestedAt),
		PaymentCompleted:       req.PaymentCompleted,
		PaymentCompletedAt:     timePtrToPgTimestamp(req.PaymentCompletedAt),
		PaymentChannelID:       uuidPtrToPgUUID(req.PaymentChannelID),
		PaymentChannelAssetID:  uuidPtrToPgUUID(req.PaymentChannelAssetID),
		AmountPaid:             int64PtrToPgInt8(req.AmountPaid),
		AmountUsd:              stringPtrToPgNumeric(req.AmountUSD),
		UpstreamUrl:            req.UpstreamURL,
		UpstreamStatusCode:     int32PtrToPgInt4(req.UpstreamStatusCode),
		UpstreamResponseTimeMs: int32PtrToPgInt4(req.UpstreamResponseTimeMs),
		FinalStatusCode:        int32PtrToPgInt4(req.FinalStatusCode),
		ErrorType:              ptrToPgText(req.ErrorType),
		ErrorMessage:           req.ErrorMessage,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toRequestLogResponse(row))
}

// ─── Request Events ───────────────────────────────────────────────────────────

// ListRequestEvents returns all events for a given request log.
// @Summary     List request events
// @Tags        request-logs
// @Produce     json
// @Param       id path string true "Request Log ID (UUID)"
// @Success     200 {array} requestEventResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/request-logs/{id}/events [get]
func (h *Handler) ListRequestEvents(c *gin.Context) {
	logID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	rows, err := h.q.ListRequestEventsByLog(c.Request.Context(), logID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]requestEventResponse, len(rows))
	for i, r := range rows {
		result[i] = toRequestEventResponse(r)
	}
	c.JSON(http.StatusOK, result)
}

type createRequestEventRequest struct {
	RequestLogID uuid.UUID       `json:"request_log_id" binding:"required"`
	EventType    string          `json:"event_type" binding:"required"`
	Metadata     json.RawMessage `json:"metadata" swaggertype:"object"`
}

// CreateRequestEvent creates a request event (internal — called by xgateway).
// @Summary     Create request event
// @Tags        request-logs
// @Accept      json
// @Produce     json
// @Param       body body createRequestEventRequest true "Event data"
// @Success     201 {object} requestEventResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    ApiKeyAuth
// @Router      /api/v1/request-events [post]
func (h *Handler) CreateRequestEvent(c *gin.Context) {
	var req createRequestEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.q.CreateRequestEvent(c.Request.Context(), postgres.CreateRequestEventParams{
		ID:           uuid.New(),
		RequestLogID: req.RequestLogID,
		EventType:    req.EventType,
		Metadata:     []byte(req.Metadata),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toRequestEventResponse(row))
}

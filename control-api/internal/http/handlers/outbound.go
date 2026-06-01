package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func ptrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func boolPtrToPgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// normalizeBazaarJSON validates that the payload is a JSON object (or empty) and
// returns the canonical bytes for storage. Returns nil bytes when the payload is empty.
func normalizeBazaarJSON(raw json.RawMessage) ([]byte, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}
	var probe map[string]any
	if err := json.Unmarshal([]byte(trimmed), &probe); err != nil {
		return nil, err
	}
	return []byte(trimmed), nil
}

// ─── Outbound Routes ─────────────────────────────────────────────────────────

type outboundRouteResponse struct {
	ID          uuid.UUID       `json:"id"`
	ProjectID   uuid.UUID       `json:"project_id"`
	ProjectName string          `json:"project_name"`
	ProjectSlug string          `json:"project_slug"`
	Name        string          `json:"name"`
	PathPattern string          `json:"path_pattern"`
	PriceUSD    string          `json:"price_usd"`
	Description string          `json:"description"`
	Free        bool            `json:"free"`
	Bazaar      json.RawMessage `json:"bazaar,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func routeListRowToResponse(r postgres.ListOutboundRoutesRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID, ProjectSlug: r.ProjectSlug,
		Name: r.Name, PathPattern: r.PathPattern,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		Bazaar:    json.RawMessage(r.Bazaar),
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeListByProjectRowToResponse(r postgres.ListOutboundRoutesByProjectRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID, ProjectSlug: r.ProjectSlug,
		Name: r.Name, PathPattern: r.PathPattern,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		Bazaar:    json.RawMessage(r.Bazaar),
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeGetRowToResponse(r postgres.GetOutboundRouteRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID,
		Name: r.Name, PathPattern: r.PathPattern,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		Bazaar:    json.RawMessage(r.Bazaar),
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeCreateRowToResponse(r postgres.CreateOutboundRouteRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID,
		Name: r.Name, PathPattern: r.PathPattern,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		Bazaar:    json.RawMessage(r.Bazaar),
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeUpdateRowToResponse(r postgres.UpdateOutboundRouteRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID,
		Name: r.Name, PathPattern: r.PathPattern,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		Bazaar:    json.RawMessage(r.Bazaar),
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

// ListOutboundRoutes returns all outbound routes, optionally filtered by project.
// @Summary     List outbound routes
// @Tags        outbound-routes
// @Produce     json
// @Param       project_id query string false "Filter by Project ID (UUID)"
// @Success     200 {array} outboundRouteResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/outbound-routes [get]
func (h *Handler) ListOutboundRoutes(c *gin.Context) {
	projects, _ := h.q.ListProjects(c.Request.Context())
	projectNames := make(map[uuid.UUID]string, len(projects))
	for _, p := range projects {
		projectNames[p.ID] = p.Name
	}

	if pid := c.Query("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		rows, err := h.q.ListOutboundRoutesByProject(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result := make([]outboundRouteResponse, len(rows))
		for i, r := range rows {
			resp := routeListByProjectRowToResponse(r)
			resp.ProjectName = projectNames[r.ProjectID]
			result[i] = resp
		}
		c.JSON(http.StatusOK, result)
		return
	}

	rows, err := h.q.ListOutboundRoutes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]outboundRouteResponse, len(rows))
	for i, r := range rows {
		resp := routeListRowToResponse(r)
		resp.ProjectName = projectNames[r.ProjectID]
		result[i] = resp
	}
	c.JSON(http.StatusOK, result)
}

// GetOutboundRoute returns an outbound route by ID.
// @Summary     Get outbound route
// @Tags        outbound-routes
// @Produce     json
// @Param       id path string true "Route ID (UUID)"
// @Success     200 {object} outboundRouteResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/outbound-routes/{id} [get]
func (h *Handler) GetOutboundRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row, err := h.q.GetOutboundRoute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, routeGetRowToResponse(row))
}

type createOutboundRouteRequest struct {
	ProjectID   uuid.UUID       `json:"project_id" binding:"required"`
	Name        string          `json:"name" binding:"required"`
	PathPattern string          `json:"path_pattern" binding:"required"`
	PriceUSD    string          `json:"price_usd"`
	Description string          `json:"description"`
	Free        bool            `json:"free"`
	Bazaar      json.RawMessage `json:"bazaar,omitempty"`
}

// CreateOutboundRoute creates a new outbound route.
// @Summary     Create outbound route
// @Tags        outbound-routes
// @Accept      json
// @Produce     json
// @Param       body body createOutboundRouteRequest true "Route data"
// @Success     201 {object} outboundRouteResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/outbound-routes [post]
// normalizePathPattern ensures the pattern has exactly one leading slash.
func normalizePathPattern(p string) string {
	return "/" + strings.TrimLeft(p, "/")
}

func (h *Handler) CreateOutboundRoute(c *gin.Context) {
	var req createOutboundRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.requireProjectOwner(c, req.ProjectID) {
		return
	}
	bazaarJSON, err := normalizeBazaarJSON(req.Bazaar)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bazaar payload: " + err.Error()})
		return
	}
	row, err := h.q.CreateOutboundRoute(c.Request.Context(), postgres.CreateOutboundRouteParams{
		ID:          uuid.New(),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		PathPattern: normalizePathPattern(req.PathPattern),
		PriceUsd:    req.PriceUSD,
		Description: req.Description,
		Free:        req.Free,
		Bazaar:      bazaarJSON,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, routeCreateRowToResponse(row))
}

type updateOutboundRouteRequest struct {
	ProjectID   *uuid.UUID      `json:"project_id"`
	Name        *string         `json:"name"`
	PathPattern *string         `json:"path_pattern"`
	PriceUSD    *string         `json:"price_usd"`
	Description *string         `json:"description"`
	Free        *bool           `json:"free"`
	Bazaar      json.RawMessage `json:"bazaar,omitempty"`
}

// UpdateOutboundRoute updates an outbound route by ID.
// @Summary     Update outbound route
// @Tags        outbound-routes
// @Accept      json
// @Produce     json
// @Param       id path string true "Route ID (UUID)"
// @Param       body body updateOutboundRouteRequest true "Fields to update"
// @Success     200 {object} outboundRouteResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/outbound-routes/{id} [put]
func (h *Handler) UpdateOutboundRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	existing, err := h.q.GetOutboundRoute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !h.requireProjectOwner(c, existing.ProjectID) {
		return
	}
	var req updateOutboundRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ProjectID != nil && *req.ProjectID != existing.ProjectID {
		if !h.requireProjectOwner(c, *req.ProjectID) {
			return
		}
	}
	if req.PathPattern != nil {
		normalized := normalizePathPattern(*req.PathPattern)
		req.PathPattern = &normalized
	}
	bazaarJSON, err := normalizeBazaarJSON(req.Bazaar)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bazaar payload: " + err.Error()})
		return
	}
	row, err := h.q.UpdateOutboundRoute(c.Request.Context(), postgres.UpdateOutboundRouteParams{
		ID:          id,
		ProjectID:   req.ProjectID,
		Name:        ptrToPgText(req.Name),
		PathPattern: ptrToPgText(req.PathPattern),
		PriceUsd:    req.PriceUSD,
		Description: req.Description,
		Free:        boolPtrToPgBool(req.Free),
		Bazaar:      bazaarJSON,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, routeUpdateRowToResponse(row))
}

// DeleteOutboundRoute deletes an outbound route by ID.
// @Summary     Delete outbound route
// @Tags        outbound-routes
// @Param       id path string true "Route ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/outbound-routes/{id} [delete]
func (h *Handler) DeleteOutboundRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	existing, err := h.q.GetOutboundRoute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !h.requireProjectOwner(c, existing.ProjectID) {
		return
	}
	ctx := c.Request.Context()
	if err := h.q.NullifyRouteInRequestLogs(ctx, &id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.q.DeleteRouteDailyStats(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.q.DeleteOutboundRoute(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

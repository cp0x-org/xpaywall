package handlers

import (
	"net/http"
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

// ─── Outbound Routes ─────────────────────────────────────────────────────────

type outboundRouteResponse struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	ProjectSlug string    `json:"project_slug"`
	Name        string    `json:"name"`
	PathPattern string    `json:"path_pattern"`
	PriceAmount int32     `json:"price_amount"`
	PriceUSD    string    `json:"price_usd"`
	Description string    `json:"description"`
	Free        bool      `json:"free"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func routeListRowToResponse(r postgres.ListOutboundRoutesRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID, ProjectSlug: r.ProjectSlug,
		Name: r.Name, PathPattern: r.PathPattern, PriceAmount: r.PriceAmount,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeListByProjectRowToResponse(r postgres.ListOutboundRoutesByProjectRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID, ProjectSlug: r.ProjectSlug,
		Name: r.Name, PathPattern: r.PathPattern, PriceAmount: r.PriceAmount,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeGetRowToResponse(r postgres.GetOutboundRouteRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID,
		Name: r.Name, PathPattern: r.PathPattern, PriceAmount: r.PriceAmount,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeCreateRowToResponse(r postgres.CreateOutboundRouteRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID,
		Name: r.Name, PathPattern: r.PathPattern, PriceAmount: r.PriceAmount,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

func routeUpdateRowToResponse(r postgres.UpdateOutboundRouteRow) outboundRouteResponse {
	return outboundRouteResponse{
		ID: r.ID, ProjectID: r.ProjectID,
		Name: r.Name, PathPattern: r.PathPattern, PriceAmount: r.PriceAmount,
		PriceUSD: r.PriceUsd, Description: r.Description, Free: r.Free,
		CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
	}
}

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
	ProjectID   uuid.UUID `json:"project_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	PathPattern string    `json:"path_pattern" binding:"required"`
	PriceAmount int32     `json:"price_amount"`
	PriceUSD    string    `json:"price_usd"`
	Description string    `json:"description"`
	Free        bool      `json:"free"`
}

func (h *Handler) CreateOutboundRoute(c *gin.Context) {
	var req createOutboundRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.q.CreateOutboundRoute(c.Request.Context(), postgres.CreateOutboundRouteParams{
		ID:          uuid.New(),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		PathPattern: req.PathPattern,
		PriceAmount: req.PriceAmount,
		PriceUsd:    req.PriceUSD,
		Description: req.Description,
		Free:        req.Free,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, routeCreateRowToResponse(row))
}

type updateOutboundRouteRequest struct {
	Name        *string `json:"name"`
	PathPattern *string `json:"path_pattern"`
	PriceAmount *int32  `json:"price_amount"`
	PriceUSD    *string `json:"price_usd"`
	Description *string `json:"description"`
	Free        *bool   `json:"free"`
}

func (h *Handler) UpdateOutboundRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateOutboundRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := h.q.UpdateOutboundRoute(c.Request.Context(), postgres.UpdateOutboundRouteParams{
		ID:          id,
		Name:        ptrToPgText(req.Name),
		PathPattern: ptrToPgText(req.PathPattern),
		PriceAmount: int32PtrToPgInt4(req.PriceAmount),
		PriceUsd:    req.PriceUSD,
		Description: req.Description,
		Free:        boolPtrToPgBool(req.Free),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, routeUpdateRowToResponse(row))
}

func (h *Handler) DeleteOutboundRoute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeleteOutboundRoute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Project Payment Configs (stubs) ─────────────────────────────────────────

func (h *Handler) ListProjectPaymentConfigs(c *gin.Context)  {}
func (h *Handler) GetProjectPaymentConfig(c *gin.Context)    {}
func (h *Handler) CreateProjectPaymentConfig(c *gin.Context) {}
func (h *Handler) UpdateProjectPaymentConfig(c *gin.Context) {}
func (h *Handler) DeleteProjectPaymentConfig(c *gin.Context) {}

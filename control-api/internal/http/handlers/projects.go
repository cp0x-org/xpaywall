package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type projectResponse struct {
	ID            uuid.UUID `json:"id"`
	OwnerUserID   uuid.UUID `json:"owner_user_id"`
	OwnerUsername string    `json:"owner_username"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type fullProjectResponse struct {
	projectResponse
	RouteSettingsID *uuid.UUID `json:"route_settings_id,omitempty"`
	BaseURL         string     `json:"base_url,omitempty"`
	AuthHeaderName  *string    `json:"auth_header_name,omitempty"`
	AuthHeaderValue *string    `json:"auth_header_value,omitempty"`
	AllowUnmatched  bool       `json:"allow_unmatched,omitempty"`
}

func toProjectResponse(p postgres.Project) projectResponse {
	return projectResponse{
		ID:          p.ID,
		OwnerUserID: p.OwnerUserID,
		Name:        p.Name,
		Slug:        p.Slug,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}

func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.q.ListProjects(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	users, _ := h.q.ListUsers(c.Request.Context())
	usernames := make(map[uuid.UUID]string, len(users))
	for _, u := range users {
		usernames[u.ID] = u.Username
	}

	result := make([]projectResponse, len(projects))
	for i, p := range projects {
		r := toProjectResponse(p)
		r.OwnerUsername = usernames[p.OwnerUserID]
		result[i] = r
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	project, err := h.q.GetProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	c.JSON(http.StatusOK, toProjectResponse(project))
}

func (h *Handler) GetFullProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	project, err := h.q.GetProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	resp := fullProjectResponse{projectResponse: toProjectResponse(project)}

	if settings, err := h.q.GetProjectRouteSettings(c.Request.Context(), id); err == nil {
		resp.RouteSettingsID = &settings.ID
		resp.BaseURL = settings.BaseUrl
		resp.AuthHeaderName = pgTextPtr(settings.AuthHeaderName)
		resp.AuthHeaderValue = pgTextPtr(settings.AuthHeaderValue)
		resp.AllowUnmatched = settings.AllowUnmatched
	}

	c.JSON(http.StatusOK, resp)
}

type createProjectRequest struct {
	OwnerUserID     uuid.UUID `json:"owner_user_id" binding:"required"`
	Name            string    `json:"name" binding:"required"`
	Slug            string    `json:"slug" binding:"required"`
	BaseURL         string    `json:"base_url" binding:"required"`
	AuthHeaderName  *string   `json:"auth_header_name"`
	AuthHeaderValue *string   `json:"auth_header_value"`
	AllowUnmatched  bool      `json:"allow_unmatched"`
}

func (h *Handler) CreateProject(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.q.CreateProject(c.Request.Context(), postgres.CreateProjectParams{
		ID:          uuid.New(),
		OwnerUserID: req.OwnerUserID,
		Name:        req.Name,
		Slug:        req.Slug,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.q.UpsertProjectRouteSettings(c.Request.Context(), postgres.UpsertProjectRouteSettingsParams{
		ID:              uuid.New(),
		ProjectID:       project.ID,
		BaseUrl:         req.BaseURL,
		AuthHeaderName:  ptrToPgText(req.AuthHeaderName),
		AuthHeaderValue: ptrToPgText(req.AuthHeaderValue),
		AllowUnmatched:  req.AllowUnmatched,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := fullProjectResponse{
		projectResponse: toProjectResponse(project),
		RouteSettingsID: &settings.ID,
		BaseURL:         settings.BaseUrl,
		AuthHeaderName:  pgTextPtr(settings.AuthHeaderName),
		AuthHeaderValue: pgTextPtr(settings.AuthHeaderValue),
		AllowUnmatched:  settings.AllowUnmatched,
	}
	c.JSON(http.StatusCreated, resp)
}

type updateProjectRequest struct {
	Name            *string `json:"name"`
	Slug            *string `json:"slug"`
	BaseURL         string  `json:"base_url" binding:"required"`
	AuthHeaderName  *string `json:"auth_header_name"`
	AuthHeaderValue *string `json:"auth_header_value"`
	AllowUnmatched  bool    `json:"allow_unmatched"`
}

func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.q.UpdateProject(c.Request.Context(), postgres.UpdateProjectParams{
		ID:   id,
		Name: ptrToPgText(req.Name),
		Slug: ptrToPgText(req.Slug),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := fullProjectResponse{projectResponse: toProjectResponse(project)}

	settings, err := h.q.UpsertProjectRouteSettings(c.Request.Context(), postgres.UpsertProjectRouteSettingsParams{
		ID:              uuid.New(),
		ProjectID:       id,
		BaseUrl:         req.BaseURL,
		AuthHeaderName:  ptrToPgText(req.AuthHeaderName),
		AuthHeaderValue: ptrToPgText(req.AuthHeaderValue),
		AllowUnmatched:  req.AllowUnmatched,
	})
	if err == nil {
		resp.RouteSettingsID = &settings.ID
		resp.BaseURL = settings.BaseUrl
		resp.AuthHeaderName = pgTextPtr(settings.AuthHeaderName)
		resp.AuthHeaderValue = pgTextPtr(settings.AuthHeaderValue)
		resp.AllowUnmatched = settings.AllowUnmatched
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeleteProject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type projectWithConfigResponse struct {
	ID             uuid.UUID `json:"id"`
	OwnerUserID    uuid.UUID `json:"owner_user_id"`
	OwnerUsername  string    `json:"owner_username"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Enabled        bool      `json:"enabled"`
	BaseURL        *string   `json:"base_url"`
	PaymentMethods []string  `json:"payment_methods"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (h *Handler) ListProjectsWithConfig(c *gin.Context) {
	projects, err := h.q.ListProjectsWithConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	users, _ := h.q.ListUsers(c.Request.Context())
	usernames := make(map[uuid.UUID]string, len(users))
	for _, u := range users {
		usernames[u.ID] = u.Username
	}

	result := make([]projectWithConfigResponse, len(projects))
	for i, p := range projects {
		result[i] = projectWithConfigResponse{
			ID:             p.ID,
			OwnerUserID:    p.OwnerUserID,
			OwnerUsername:  usernames[p.OwnerUserID],
			Name:           p.Name,
			Slug:           p.Slug,
			Enabled:        p.Enabled,
			BaseURL:        pgTextPtr(p.BaseUrl),
			PaymentMethods: toStringSlice(p.PaymentMethods),
			CreatedAt:      p.CreatedAt.Time,
			UpdatedAt:      p.UpdatedAt.Time,
		}
	}
	c.JSON(http.StatusOK, result)
}

func toStringSlice(v any) []string {
	if v == nil {
		return []string{}
	}
	arr, ok := v.([]any)
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func (h *Handler) GetProjectSettings(c *gin.Context)    {}
func (h *Handler) UpdateProjectSettings(c *gin.Context) {}

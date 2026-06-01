package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type projectResponse struct {
	ID            uuid.UUID  `json:"id"`
	OwnerUserID   *uuid.UUID `json:"owner_user_id"`
	OwnerUsername string     `json:"owner_username"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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

// ListProjects returns all projects.
// @Summary     List projects
// @Tags        projects
// @Produce     json
// @Success     200 {array} projectResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects [get]
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
		if p.OwnerUserID != nil {
			r.OwnerUsername = usernames[*p.OwnerUserID]
		}
		result[i] = r
	}
	c.JSON(http.StatusOK, result)
}

// GetProject returns a project by ID.
// @Summary     Get project
// @Tags        projects
// @Produce     json
// @Param       id path string true "Project ID (UUID)"
// @Success     200 {object} projectResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects/{id} [get]
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

// GetFullProject returns a project with its route settings.
// @Summary     Get project with settings
// @Tags        projects
// @Produce     json
// @Param       id path string true "Project ID (UUID)"
// @Success     200 {object} fullProjectResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects/{id}/full [get]
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

	if resp.OwnerUserID != nil {
		if user, err := h.q.GetUser(c.Request.Context(), *resp.OwnerUserID); err == nil {
			resp.OwnerUsername = user.Username
		}
	}

	c.JSON(http.StatusOK, resp)
}

type createProjectRequest struct {
	Name            string  `json:"name" binding:"required"`
	Slug            string  `json:"slug" binding:"required"`
	BaseURL         string  `json:"base_url" binding:"required"`
	AuthHeaderName  *string `json:"auth_header_name"`
	AuthHeaderValue *string `json:"auth_header_value"`
	AllowUnmatched  bool    `json:"allow_unmatched"`
}

// CreateProject creates a new project with route settings.
// @Summary     Create project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       body body createProjectRequest true "Project data"
// @Success     201 {object} fullProjectResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	callerID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id in token"})
		return
	}
	ownerID, ok := callerID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id type"})
		return
	}

	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.q.CreateProject(c.Request.Context(), postgres.CreateProjectParams{
		ID:          uuid.New(),
		OwnerUserID: &ownerID,
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

// UpdateProject updates a project and its route settings.
// @Summary     Update project
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       id path string true "Project ID (UUID)"
// @Param       body body updateProjectRequest true "Fields to update"
// @Success     200 {object} fullProjectResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects/{id} [put]
func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if !h.requireProjectOwner(c, id) {
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

// DeleteProject archives a project by ID. Archived projects are hidden from
// listings and stop resolving in xgateway, but their referenced rows
// (request_logs, project_daily_stats, routes) remain intact.
// @Summary     Delete (archive) project
// @Tags        projects
// @Param       id path string true "Project ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects/{id} [delete]
func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if !h.requireProjectOwner(c, id) {
		return
	}
	if err := h.q.ArchiveProject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type projectWithConfigResponse struct {
	ID             uuid.UUID  `json:"id"`
	OwnerUserID    *uuid.UUID `json:"owner_user_id"`
	OwnerUsername  string     `json:"owner_username"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Enabled        bool       `json:"enabled"`
	BaseURL        *string    `json:"base_url"`
	PaymentMethods []string   `json:"payment_methods"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ListProjectsWithConfig returns all projects with their payment configuration.
// @Summary     List projects with config
// @Tags        projects
// @Produce     json
// @Success     200 {array} projectWithConfigResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/projects/with-config [get]
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
			ID:          p.ID,
			OwnerUserID: p.OwnerUserID,
			OwnerUsername: func() string {
				if p.OwnerUserID != nil {
					return usernames[*p.OwnerUserID]
				}
				return ""
			}(),
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

// GetProjectSettings returns project settings (not yet implemented).
// @Summary     Get project settings
// @Tags        projects
// @Produce     json
// @Param       projectId path string true "Project ID (UUID)"
// @Success     200 {object} object
// @Security    BearerAuth
// @Router      /api/v1/project-settings/{projectId} [get]
func (h *Handler) GetProjectSettings(c *gin.Context) {}

// UpdateProjectSettings updates project settings (not yet implemented).
// @Summary     Update project settings
// @Tags        projects
// @Accept      json
// @Produce     json
// @Param       projectId path string true "Project ID (UUID)"
// @Success     200 {object} object
// @Security    BearerAuth
// @Router      /api/v1/project-settings/{projectId} [put]
func (h *Handler) UpdateProjectSettings(c *gin.Context) {}

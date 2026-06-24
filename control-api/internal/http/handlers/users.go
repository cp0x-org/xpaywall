package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
	"github.com/cp0x-org/xpaywall/control-api/internal/validate"
)

// ListUsers returns all users.
// @Summary     List users
// @Tags        users
// @Produce     json
// @Success     200 {array} object "Array of user objects"
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.q.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUser returns a user by ID.
// @Summary     Get user
// @Tags        users
// @Produce     json
// @Param       id path string true "User ID (UUID)"
// @Success     200 {object} object "User object"
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user, err := h.q.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// CreateUser creates a new user.
// @Summary     Create user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       body body createUserRequest true "User data"
// @Success     201 {object} object "Created user object"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validate.Slug(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username may contain only letters, digits, underscore and hyphen"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := h.q.CreateUser(c.Request.Context(), postgres.CreateUserParams{
		ID:           uuid.New(),
		Username:     req.Username,
		AuthProvider: "local",
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

type updateUserRequest struct {
	Username *string `json:"username"`
}

// UpdateUser updates a user by ID.
// @Summary     Update user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id path string true "User ID (UUID)"
// @Param       body body updateUserRequest true "Fields to update"
// @Success     200 {object} object "Updated user object"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username != nil && !validate.Slug(*req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username may contain only letters, digits, underscore and hyphen"})
		return
	}

	user, err := h.q.UpdateUser(c.Request.Context(), postgres.UpdateUserParams{
		ID:       id,
		Username: ptrToPgText(req.Username),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user by ID.
// @Summary     Delete user
// @Tags        users
// @Param       id path string true "User ID (UUID)"
// @Success     204 "No Content"
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.q.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

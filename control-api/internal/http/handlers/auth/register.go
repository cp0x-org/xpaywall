package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
	"github.com/cp0x-org/xpaywall/control-api/internal/validate"
)

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register creates a local account and returns a JWT token (auto-login).
// @Summary     Register
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Registration details"
// @Success     200 {object} authResponse
// @Failure     400 {object} errorResponse
// @Failure     409 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if !validate.Slug(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username may contain only letters, digits, underscore and hyphen"})
		return
	}

	ctx := c.Request.Context()
	if _, err := h.q.GetUserByUsername(ctx, req.Username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}
	if _, err := h.q.GetUserByEmail(ctx, pgtype.Text{String: req.Email, Valid: true}); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := h.q.CreateUser(ctx, postgres.CreateUserParams{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        pgtype.Text{String: req.Email, Valid: true},
		AuthProvider: "local",
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Welcome email is best-effort: a delivery failure must not block signup.
	if err := h.mailer.SendWelcome(ctx, req.Email, req.Username); err != nil {
		log.Printf("welcome email to %s failed: %v", req.Email, err)
	}

	token, err := h.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, User: toUserResponse(user)})
}

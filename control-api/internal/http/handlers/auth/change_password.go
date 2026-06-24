package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword updates the authenticated user's password.
// @Summary     Change password
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body changePasswordRequest true "Current and new password"
// @Success     200 {object} forgotPasswordResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.q.GetUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// OAuth-only users have no password to verify; they must use forgot-password.
	if !user.PasswordHash.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account has no password set"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.q.UpdateUserPassword(ctx, postgres.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, forgotPasswordResponse{Message: "password updated"})
}

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

// passwordResetTTL bounds how long a reset link stays valid (requirement: <= 1h).
const passwordResetTTL = time.Hour

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type forgotPasswordResponse struct {
	Message string `json:"message"`
	// ResetURL is returned temporarily until SMTP delivery is wired up.
	ResetURL string `json:"reset_url,omitempty"`
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// hashResetToken returns the value stored in the DB for a raw reset token.
func hashResetToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// ForgotPassword issues a password reset link valid for one hour.
// It always returns 200 so callers cannot probe which emails are registered.
// @Summary     Request password reset
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body forgotPasswordRequest true "Account email"
// @Success     200 {object} forgotPasswordResponse
// @Failure     400 {object} errorResponse
// @Router      /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	resp := forgotPasswordResponse{Message: "if the email is registered, a reset link has been sent"}

	user, err := h.q.GetUserByEmail(ctx, pgtype.Text{String: req.Email, Valid: true})
	if err != nil {
		c.JSON(http.StatusOK, resp)
		return
	}

	// Invalidate any outstanding tokens before issuing a fresh one.
	if err := h.q.DeleteUserPasswordResetTokens(ctx, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	rawToken, err := generateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	_, err = h.q.CreatePasswordResetToken(ctx, postgres.CreatePasswordResetTokenParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashResetToken(rawToken),
		ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(passwordResetTTL), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", h.cfg.AppBaseURL, rawToken)
	if err := h.mailer.SendPasswordReset(ctx, req.Email, link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send reset email"})
		return
	}

	// When SMTP is unconfigured the email is only logged, so surface the link in
	// the response to keep the dev flow usable. With real delivery the link is
	// sent only by email and never returned (returning it would let anyone reset
	// any account).
	if !h.cfg.MailEnabled() {
		resp.ResetURL = link
	}
	c.JSON(http.StatusOK, resp)
}

// ResetPassword consumes a reset token and sets a new password.
// @Summary     Reset password
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body resetPasswordRequest true "Reset token and new password"
// @Success     200 {object} forgotPasswordResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Router      /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	token, err := h.q.GetActivePasswordResetToken(ctx, hashResetToken(req.Token))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.q.UpdateUserPassword(ctx, postgres.UpdateUserPasswordParams{
		ID:           token.UserID,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	if err := h.q.MarkPasswordResetTokenUsed(ctx, token.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, forgotPasswordResponse{Message: "password updated"})
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

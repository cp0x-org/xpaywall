package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/idtoken"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
	"github.com/cp0x-org/xpaywall/control-api/internal/validate"
)

type googleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// GoogleAuth signs in (or registers) a user from a Google ID token.
// The frontend obtains the token via Google Identity Services and posts it here.
// @Summary     Google sign-in
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body googleAuthRequest true "Google ID token"
// @Success     200 {object} authResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Router      /auth/google [post]
func (h *Handler) GoogleAuth(c *gin.Context) {
	var req googleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if h.cfg.GoogleClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "google login not configured"})
		return
	}

	ctx := c.Request.Context()
	payload, err := idtoken.Validate(ctx, req.IDToken, h.cfg.GoogleClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid google token"})
		return
	}

	googleID := payload.Subject
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	if googleID == "" || email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid google token"})
		return
	}

	user, err := h.resolveGoogleUser(ctx, googleID, email, name, picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign in with google"})
		return
	}

	token, err := h.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, User: toUserResponse(user)})
}

// resolveGoogleUser implements find-or-link-or-create:
//  1. known google_id  -> sign in
//  2. same email       -> link Google to the existing account
//  3. otherwise        -> create a new Google-backed account
func (h *Handler) resolveGoogleUser(ctx context.Context, googleID, email, name, picture string) (postgres.User, error) {
	emailText := pgtype.Text{String: email, Valid: true}
	googleIDText := pgtype.Text{String: googleID, Valid: true}
	avatar := pgtype.Text{String: picture, Valid: picture != ""}

	if user, err := h.q.GetUserByGoogleID(ctx, googleIDText); err == nil {
		return user, nil
	}

	if existing, err := h.q.GetUserByEmail(ctx, emailText); err == nil {
		return h.q.LinkGoogleAccount(ctx, postgres.LinkGoogleAccountParams{
			ID:        existing.ID,
			GoogleID:  googleIDText,
			AvatarUrl: avatar,
		})
	}

	return h.q.CreateGoogleUser(ctx, postgres.CreateGoogleUserParams{
		ID:        uuid.New(),
		Username:  h.uniqueUsername(ctx, usernameFromGoogle(name, email)),
		Email:     emailText,
		GoogleID:  googleIDText,
		AvatarUrl: avatar,
	})
}

// uniqueUsername returns base, or base with a short random suffix on collision.
func (h *Handler) uniqueUsername(ctx context.Context, base string) string {
	if _, err := h.q.GetUserByUsername(ctx, base); err != nil {
		return base
	}
	return base + "-" + uuid.New().String()[:8]
}

// usernameFromGoogle derives a slug-safe username base from the Google profile.
// The name/email may contain spaces, dots or other characters, so the result is
// sanitized; if nothing usable remains it falls back to "user" (uniqueUsername
// then appends a random suffix on collision).
func usernameFromGoogle(name, email string) string {
	base := name
	if base == "" {
		for i, r := range email {
			if r == '@' {
				base = email[:i]
				break
			}
		}
		if base == "" {
			base = email
		}
	}
	if slug := validate.Sanitize(base); slug != "" {
		return slug
	}
	return "user"
}

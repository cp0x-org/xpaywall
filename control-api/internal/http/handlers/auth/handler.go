package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/cp0x-org/xpaywall/control-api/config"
	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type Handler struct {
	cfg       *config.ControlAPIConfig
	q         *postgres.Queries
	mailer    Mailer
	jwtSecret string
}

func New(cfg *config.ControlAPIConfig, q *postgres.Queries, mailer Mailer) *Handler {
	return &Handler{
		cfg:       cfg,
		q:         q,
		mailer:    mailer,
		jwtSecret: cfg.JWTSecret,
	}
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email,omitempty"`
	Role     string    `json:"role"`
}

func toUserResponse(u postgres.User) userResponse {
	return userResponse{ID: u.ID, Username: u.Username, Email: u.Email.String, Role: u.Role}
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Login authenticates a user and returns a JWT token.
// @Summary     Login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Login credentials"
// @Success     200 {object} authResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Router      /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// The login identifier may be either a username or an email address.
	user, err := h.q.GetUserByUsernameOrEmail(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// OAuth-only users (e.g. Google) have no password hash and cannot password-login.
	if !user.PasswordHash.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	token, err := h.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token: token,
		User:  toUserResponse(user),
	})
}

// Me returns the currently authenticated user.
// @Summary     Get current user
// @Tags        auth
// @Produce     json
// @Success     200 {object} userResponse
// @Failure     401 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Security    BearerAuth
// @Router      /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
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

	user, err := h.q.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *Handler) generateToken(userID uuid.UUID, username, role string) (string, error) {
	cl := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString([]byte(h.jwtSecret))
}

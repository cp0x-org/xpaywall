package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SuperadminOnly aborts the request unless the caller is the superadmin.
// Must be installed after JWT so user_id is present in the gin context.
// Superadmin is identified by user_id == uuid.Nil (see auth handler login flow).
func SuperadminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		id, ok := v.(uuid.UUID)
		if !ok || id != uuid.Nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "superadmin only"})
			return
		}
		c.Next()
	}
}

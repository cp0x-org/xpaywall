package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoleSuperadmin is the users.role value granting global-entity management rights.
const RoleSuperadmin = "superadmin"

// SuperadminOnly aborts the request unless the caller has the superadmin role.
// Must be installed after JWT so role is present in the gin context.
// Superadmin is provisioned directly in Postgres (UPDATE users SET role='superadmin').
func SuperadminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		role, ok := v.(string)
		if !ok || role != RoleSuperadmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "superadmin only"})
			return
		}
		c.Next()
	}
}

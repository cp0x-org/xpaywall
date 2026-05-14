package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const internalAPIKeyHeader = "X-Api-Key"

func InternalAPIKey(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(internalAPIKeyHeader) != key {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

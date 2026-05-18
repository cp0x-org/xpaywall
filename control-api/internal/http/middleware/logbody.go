package middleware

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
)

// LogRequestBody logs the raw JSON body for debugging, then restores it for the handler.
func LogRequestBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err == nil {
			log.Printf("[body] %s %s: %s", c.Request.Method, c.Request.URL.Path, body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		c.Next()
	}
}

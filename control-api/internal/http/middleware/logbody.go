package middleware

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

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

// DebugLog logs request and response bodies to stdout.
func DebugLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(reqBody))

		rbw := &responseBodyWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = rbw

		c.Next()

		status := c.Writer.Status()
		log.Printf("[debug] --> %s %s\n  req body: %s", c.Request.Method, c.Request.URL.Path, reqBody)
		log.Printf("[debug] <-- %s %s  status=%d\n  res body: %s",
			c.Request.Method, c.Request.URL.Path, status, rbw.body.String())
	}
}

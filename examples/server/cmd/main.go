package main

import (
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// demoBearerToken is the bearer token the protected demo endpoints require in
// the Authorization header. It mirrors what an operator would put in a route's
// auth_header_value (e.g. "Bearer <token>") so xgateway injects it on the
// upstream request. Override with DEMO_BEARER_TOKEN.
func demoBearerToken() string {
	if t := os.Getenv("DEMO_BEARER_TOKEN"); t != "" {
		return t
	}
	return "demo-secret-token"
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(corsMiddleware)

	// x402 demo endpoints (seeded by `install demo`).
	r.GET("/health", handleHealth)
	r.Any("/api/metered/*path", handleMetered)
	r.GET("/weather", handleWeather)
	r.GET("/free-endpoint", freeEndpoint)
	r.GET("/free-multipoint/*path", freeMultipoint)
	r.GET("/http-endpoint", httpEndpoint)

	// MPP demo endpoints (seeded by `install demo-mpp`); kept disjoint from the
	// x402 set above so the two demos can be tested without route overlap.
	r.GET("/time", handleTime)
	r.Any("/api/usage/*path", handleUsage)
	r.GET("/quote", handleQuote)
	r.GET("/ping", handlePing)
	r.GET("/echo/*path", handleEcho)

	// Auth-protected demo endpoints: require a valid bearer token in the
	// Authorization header. Configure a route in the admin panel with
	// auth_header_name="Authorization" and auth_header_value="Bearer <token>"
	// so xgateway injects the credential after payment is verified. A direct
	// request without the header gets 401 -- proving the upstream is reachable
	// only through the gateway.
	protected := r.Group("/protected", bearerAuthMiddleware)
	protected.GET("", handleProtected)
	protected.GET("/*path", handleProtected)

	r.Run(":4021")
}

// bearerAuthMiddleware rejects requests that don't carry the expected
// "Authorization: Bearer <token>" header with 401.
func bearerAuthMiddleware(c *gin.Context) {
	const prefix = "Bearer "
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, prefix) || strings.TrimSpace(auth[len(prefix):]) != demoBearerToken() {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":  "missing or invalid bearer token",
			"status": http.StatusUnauthorized,
		})
		return
	}
	c.Next()
}

func handleProtected(c *gin.Context) {
	path := c.Param("path")
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"path":      "/protected" + path,
		"data":      "authorized: bearer token accepted",
		"timestamp": time.Now().UTC(),
	})
}

// corsMiddleware sets permissive CORS headers so the demo server can be called
// directly from a browser (e.g. the admin panel's Bazaar auto-generator).
func corsMiddleware(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin == "" {
		origin = "*"
	}
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Vary", "Origin")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Api-Key, X-Requested-With")
	c.Header("Access-Control-Max-Age", "86400")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func httpEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"data":      "http endpoint",
		"timestamp": time.Now().UTC(),
	})
}

func handleMetered(c *gin.Context) {
	path := c.Param("path")
	units := rand.Intn(10) + 1
	c.JSON(http.StatusOK, gin.H{
		"path":       "/api/metered" + path,
		"units_used": units,
		"cost_usd":   float64(units) * 0.01,
		"timestamp":  time.Now().UTC(),
		"data":       "mock metered response",
	})
}

func handleWeather(c *gin.Context) {
	cities := []string{"Kyiv", "New York", "London", "Tokyo"}
	city := c.Query("city")
	if city == "" {
		city = cities[rand.Intn(len(cities))]
	}
	c.JSON(http.StatusOK, gin.H{
		"city":        city,
		"temperature": rand.Intn(35) - 5,
		"unit":        "celsius",
		"condition":   []string{"sunny", "cloudy", "rainy", "windy"}[rand.Intn(4)],
		"humidity":    rand.Intn(60) + 30,
		"timestamp":   time.Now().UTC(),
	})
}

func freeEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"data":      "free endpoint",
		"timestamp": time.Now().UTC(),
	})
}

func freeMultipoint(c *gin.Context) {
	path := c.Param("path")
	c.JSON(http.StatusOK, gin.H{
		"path":      "/free-multipoint" + path,
		"data":      "free multipoint",
		"timestamp": time.Now().UTC(),
	})
}

// --- MPP demo endpoints ---

func handleTime(c *gin.Context) {
	now := time.Now().UTC()
	c.JSON(http.StatusOK, gin.H{
		"timestamp": now,
		"unix":      now.Unix(),
		"timezone":  "UTC",
	})
}

func handleUsage(c *gin.Context) {
	path := c.Param("path")
	units := rand.Intn(20) + 1
	c.JSON(http.StatusOK, gin.H{
		"path":       "/api/usage" + path,
		"units_used": units,
		"cost_usd":   float64(units) * 0.005,
		"timestamp":  time.Now().UTC(),
		"data":       "mock usage response",
	})
}

func handleQuote(c *gin.Context) {
	symbols := []string{"BTC", "ETH", "SOL", "TEMPO"}
	symbol := c.Query("symbol")
	if symbol == "" {
		symbol = symbols[rand.Intn(len(symbols))]
	}
	c.JSON(http.StatusOK, gin.H{
		"symbol":    symbol,
		"price_usd": float64(rand.Intn(50000)) + rand.Float64(),
		"currency":  "USD",
		"timestamp": time.Now().UTC(),
	})
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"data":      "pong",
		"timestamp": time.Now().UTC(),
	})
}

func handleEcho(c *gin.Context) {
	path := c.Param("path")
	c.JSON(http.StatusOK, gin.H{
		"path":      "/echo" + path,
		"data":      "echo",
		"timestamp": time.Now().UTC(),
	})
}

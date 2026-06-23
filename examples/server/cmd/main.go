package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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

	r.Run(":4021")
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

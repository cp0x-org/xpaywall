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

	r.GET("/health", handleHealth)
	r.Any("/api/metered/*path", handleMetered)
	r.GET("/weather", handleWeather)
	r.GET("/free-endpoint", freeEndpoint)
	r.GET("/free-multipoint/*path", freeMultipoint)
	r.GET("/http-endpoint", httpEndpoint)

	r.Run(":4021")
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

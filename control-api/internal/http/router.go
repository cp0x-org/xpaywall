package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers"
	authhandler "github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/auth"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/gateway"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/middleware"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/routes"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func SetupRouter(h *handlers.Handler, authHandler *authhandler.Handler, gatewayHandler *gateway.Handler, internalAPIKey, jwtSecret string) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware())

	// Public auth endpoints
	auth := router.Group("/auth")
	auth.POST("/login", authHandler.Login)

	// JWT-protected routes
	protected := router.Group("/")
	protected.Use(middleware.JWT(jwtSecret))
	protected.GET("/auth/me", authHandler.Me)
	routes.RegisterAdminRoutes(protected, h)

	// Internal proxy endpoints (gateway → control-api)
	proxy := router.Group("/")
	proxy.Use(middleware.InternalAPIKey(internalAPIKey))
	routes.RegisterProxyRoutes(proxy, gatewayHandler)
	routes.RegisterInternalRoutes(proxy, h)

	return router
}

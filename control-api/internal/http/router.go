package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/cp0x-org/xpaywall/control-api/docs"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers"
	authhandler "github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/auth"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/gateway"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/middleware"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/routes"
)

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowAny := false
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			allowAny = true
			continue
		}
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		echoed := false
		switch {
		case allowAny && origin == "":
			c.Header("Access-Control-Allow-Origin", "*")
		case allowAny:
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			echoed = true
		default:
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				echoed = true
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		reqHeaders := c.GetHeader("Access-Control-Request-Headers")
		if reqHeaders != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		}
		c.Header("Access-Control-Max-Age", "600")
		if echoed {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func SetupRouter(h *handlers.Handler, authHandler *authhandler.Handler, gatewayHandler *gateway.Handler, internalAPIKey, jwtSecret string, debug bool, corsOrigins []string) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware(corsOrigins))
	// Gin skips global middleware on unmatched routes unless NoRoute is set.
	// This registration ensures CORS preflight (OPTIONS) on any path runs the middleware.
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})
	if debug {
		router.Use(middleware.DebugLog())
	}

	// Swagger UI — available at /swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	proxy.Use(middleware.LogRequestBody())
	routes.RegisterProxyRoutes(proxy, gatewayHandler)
	routes.RegisterInternalRoutes(proxy, h)

	return router
}

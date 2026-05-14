package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/gateway"
)

func RegisterProxyRoutes(router gin.IRouter, h *gateway.Handler) {
	proxy := router.Group("/proxy")
	proxy.GET("/resolve/*path", h.ResolveRoute)
}

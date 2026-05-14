package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers"
)

func RegisterInternalRoutes(router gin.IRouter, h *handlers.Handler) {
	api := router.Group("/api/v1")

	logs := api.Group("/request-logs")
	logs.POST("", h.CreateRequestLog)
	logs.PUT("/:id", h.UpdateRequestLog)

	events := api.Group("/request-events")
	events.POST("", h.CreateRequestEvent)
}

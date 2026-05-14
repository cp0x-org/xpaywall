package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers"
)

func RegisterAdminRoutes(router gin.IRouter, h *handlers.Handler) {
	api := router.Group("/api/v1")

	registerLogsRoutes(api, h)
	registerSystemRoutes(api, h)
	registerStatsRoutes(api, h)
	registerUserRoutes(api, h)
	registerProjectRoutes(api, h)
	registerPaymentChannelRoutes(api, h)
	registerPaymentChannelAssetRoutes(api, h)

	registerOutboundRouteRoutes(api, h)

	registerProjectSettingsRoutes(api, h)
	registerProjectPaymentConfigRoutes(api, h)
}

func registerSystemRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	users := api.Group("/system")
	users.GET("/proxy-url", h.GetProxyUrl)
}

func registerStatsRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	stats := api.Group("/stats")
	stats.GET("/dashboard", h.GetDashboardStats)
	stats.GET("/daily", h.GetDailyStats)
	stats.GET("/recent-requests", h.GetDashboardRecentRequests)
	stats.GET("/top-routes", h.GetDashboardTopRoutes)
}

func registerUserRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	users := api.Group("/users")
	users.GET("", h.ListUsers)
	users.GET("/:id", h.GetUser)
	users.POST("", h.CreateUser)
	users.PUT("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeleteUser)
}

func registerProjectRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	projects := api.Group("/projects")
	projects.GET("", h.ListProjects)
	projects.GET("/with-config", h.ListProjectsWithConfig)
	projects.GET("/:id", h.GetProject)
	projects.GET("/:id/full", h.GetFullProject)
	projects.POST("", h.CreateProject)
	projects.PUT("/:id", h.UpdateProject)
	projects.DELETE("/:id", h.DeleteProject)
}

func registerProjectSettingsRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	projectSettings := api.Group("/project-settings")
	projectSettings.GET("/:projectId", h.GetProjectSettings)
	projectSettings.PUT("/:projectId", h.UpdateProjectSettings)
}

func registerOutboundRouteRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	outboundRoutes := api.Group("/outbound-routes")
	outboundRoutes.GET("", h.ListOutboundRoutes)
	outboundRoutes.GET("/:id", h.GetOutboundRoute)
	outboundRoutes.POST("", h.CreateOutboundRoute)
	outboundRoutes.PUT("/:id", h.UpdateOutboundRoute)
	outboundRoutes.DELETE("/:id", h.DeleteOutboundRoute)
}

func registerPaymentChannelRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	paymentChannels := api.Group("/payment-channels")
	paymentChannels.GET("", h.ListPaymentChannels)
	paymentChannels.GET("/:id", h.GetPaymentChannel)
	paymentChannels.POST("", h.CreatePaymentChannel)
	paymentChannels.PUT("/:id", h.UpdatePaymentChannel)
	paymentChannels.DELETE("/:id", h.DeletePaymentChannel)
}

func registerPaymentChannelAssetRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	paymentChannelAssets := api.Group("/payment-channel-assets")
	paymentChannelAssets.GET("", h.ListPaymentChannelAssets)
	paymentChannelAssets.GET("/:id", h.GetPaymentChannelAsset)
	paymentChannelAssets.POST("", h.CreatePaymentChannelAsset)
	paymentChannelAssets.PUT("/:id", h.UpdatePaymentChannelAsset)
	paymentChannelAssets.DELETE("/:id", h.DeletePaymentChannelAsset)
}

func registerProjectPaymentConfigRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	projectPaymentConfigs := api.Group("/project-payment-configs")
	projectPaymentConfigs.GET("", h.ListProjectPaymentConfigs)
	projectPaymentConfigs.GET("/:id", h.GetProjectPaymentConfig)
	projectPaymentConfigs.POST("", h.CreateProjectPaymentConfig)
	projectPaymentConfigs.PUT("/:id", h.UpdateProjectPaymentConfig)
	projectPaymentConfigs.DELETE("/:id", h.DeleteProjectPaymentConfig)
}

func registerLogsRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	logs := api.Group("/request-logs")
	logs.GET("", h.ListRequestLogs)
	logs.GET("/:id", h.GetRequestLog)
	logs.GET("/:id/events", h.ListRequestEvents)
}

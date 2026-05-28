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
	registerPaymentMethodRoutes(api, h)
	registerPaymentMethodAssetRoutes(api, h)
	registerFacilitatorRoutes(api, h)

	registerOutboundRouteRoutes(api, h)

	registerProjectSettingsRoutes(api, h)
	registerProjectPaymentMethodRoutes(api, h)
}

func registerSystemRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	users := api.Group("/system")
	users.GET("/proxy-url", h.GetProxyUrl)
	users.GET("/networks", h.ListNetworks)
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

func registerPaymentMethodRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	paymentMethods := api.Group("/payment-methods")
	paymentMethods.GET("", h.ListPaymentMethods)
	paymentMethods.GET("/:id", h.GetPaymentMethod)
	paymentMethods.POST("", h.CreatePaymentMethod)
	paymentMethods.PUT("/:id", h.UpdatePaymentMethod)
	paymentMethods.DELETE("/:id", h.DeletePaymentMethod)
}

func registerPaymentMethodAssetRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	paymentMethodAssets := api.Group("/payment-method-assets")
	paymentMethodAssets.GET("", h.ListPaymentMethodAssets)
	paymentMethodAssets.GET("/:id", h.GetPaymentMethodAsset)
	paymentMethodAssets.POST("", h.CreatePaymentMethodAsset)
	paymentMethodAssets.PUT("/:id", h.UpdatePaymentMethodAsset)
	paymentMethodAssets.DELETE("/:id", h.DeletePaymentMethodAsset)
}

func registerFacilitatorRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	facilitators := api.Group("/facilitators")
	facilitators.GET("", h.ListFacilitators)
	facilitators.GET("/:id", h.GetFacilitator)
	facilitators.POST("", h.CreateFacilitator)
	facilitators.PUT("/:id", h.UpdateFacilitator)
	facilitators.DELETE("/:id", h.DeleteFacilitator)
}

func registerProjectPaymentMethodRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	projectPaymentMethods := api.Group("/project-payment-methods")
	projectPaymentMethods.GET("", h.ListProjectPaymentMethods)
	projectPaymentMethods.GET("/all", h.ListAllProjectPaymentMethods)
	projectPaymentMethods.GET("/:id", h.GetProjectPaymentMethod)
	projectPaymentMethods.POST("", h.CreateProjectPaymentMethod)
	projectPaymentMethods.PUT("/:id", h.UpdateProjectPaymentMethod)
	projectPaymentMethods.DELETE("/:id", h.DeleteProjectPaymentMethod)
}

func registerLogsRoutes(api *gin.RouterGroup, h *handlers.Handler) {
	logs := api.Group("/request-logs")
	logs.GET("", h.ListRequestLogs)
	logs.GET("/:id", h.GetRequestLog)
	logs.GET("/:id/events", h.ListRequestEvents)
}

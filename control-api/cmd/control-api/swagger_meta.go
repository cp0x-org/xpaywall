// @title           xpaywall Control API
// @version         1.0
// @description     Control plane for xpaywall — manages projects, routes, users, payment channels, and request logs.
// @host            localhost:9090
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token. Format: "Bearer {token}"

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Api-Key
// @description Internal API key shared between xgateway and control-api
package main

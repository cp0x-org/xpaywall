package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type proxyResponse struct {
	ProxyUrl string `json:"proxy_url"`
}

// GetProxyUrl returns the public xgateway proxy URL.
// @Summary     Get proxy URL
// @Tags        system
// @Produce     json
// @Success     200 {object} proxyResponse
// @Failure     401 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/system/proxy-url [get]
func (h *Handler) GetProxyUrl(c *gin.Context) {
	c.JSON(http.StatusOK, &proxyResponse{
		ProxyUrl: h.cfg.ProxyUrl,
	})
}

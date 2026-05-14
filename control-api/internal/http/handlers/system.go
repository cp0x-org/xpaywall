package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type proxyResponse struct {
	ProxyUrl string `json:"proxy_url"`
}

func (h *Handler) GetProxyUrl(c *gin.Context) {
	c.JSON(http.StatusOK, &proxyResponse{
		ProxyUrl: h.cfg.ProxyUrl,
	})
}

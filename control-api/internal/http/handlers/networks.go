package handlers

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/cp0x-org/xpaywall/control-api/internal/networks"
)

type networkItem struct {
	CAIP2 string `json:"caip2"`
	Name  string `json:"name"`
}

func (h *Handler) ListNetworks(c *gin.Context) {
	items := make([]networkItem, 0, len(networks.ByCAIP2))
	for caip2, name := range networks.ByCAIP2 {
		items = append(items, networkItem{CAIP2: caip2, Name: name})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	c.JSON(http.StatusOK, items)
}

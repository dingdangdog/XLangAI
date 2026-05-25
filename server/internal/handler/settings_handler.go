package handler

import (
	"net/http"

	"wlltalk/server/internal/settings"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	sys *settings.Service
}

func NewSettingsHandler(sys *settings.Service) *SettingsHandler {
	return &SettingsHandler{sys: sys}
}

// PublicSettings 返回允许客户端读取的系统变量（无厂商密钥）。
func (h *SettingsHandler) PublicSettings(c *gin.Context) {
	m, err := h.sys.PublicMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": m})
}

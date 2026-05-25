package handler

import (
	"net/http"
	"strings"

	"wlltalk/server/config"
	"wlltalk/server/internal/auth"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	_ = cfg
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		userID, err := auth.ParseToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

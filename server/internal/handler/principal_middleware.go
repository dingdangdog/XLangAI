package handler

import (
	"net/http"

	"wlltalk/server/internal/authz"

	"github.com/gin-gonic/gin"
)

const principalCtxKey = "principal"

func PrincipalMiddleware(svc *authz.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("user_id")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		p, err := svc.LoadPrincipal(c.Request.Context(), uid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Set(principalCtxKey, p)
		c.Next()
	}
}

func CtxPrincipal(c *gin.Context) *authz.Principal {
	v, ok := c.Get(principalCtxKey)
	if !ok {
		return nil
	}
	p, _ := v.(*authz.Principal)
	return p
}

// RequireRole 管理类接口使用，例如 admin。
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := CtxPrincipal(c)
		if p == nil || p.Role != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

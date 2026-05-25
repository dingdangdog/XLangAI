package handler

import (
	"net/http"
	"time"

	"wlltalk/server/internal/cache"
	"wlltalk/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type LangHandler struct {
	repo  *repository.LangRepo
	cache *cache.Cache
	ttl   time.Duration
}

func NewLangHandler(repo *repository.LangRepo, c *cache.Cache, ttl time.Duration) *LangHandler {
	return &LangHandler{repo: repo, cache: c, ttl: ttl}
}

func (h *LangHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	if h.cache != nil && h.ttl > 0 {
		var cached []repository.Language
		if h.cache.GetJSON(ctx, cache.LangListKey(), &cached) && len(cached) > 0 {
			c.JSON(http.StatusOK, cached)
			return
		}
	}
	langs, err := h.repo.List(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.cache != nil && h.ttl > 0 {
		h.cache.SetJSON(ctx, cache.LangListKey(), langs, h.ttl)
	}
	c.JSON(http.StatusOK, langs)
}

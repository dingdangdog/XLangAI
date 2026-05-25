package handler

import (
	"errors"
	"net/http"
	"strings"

	"xlangai/server/internal/media"
	"xlangai/server/internal/objectstore"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	media *media.Service
}

func NewMediaHandler(mediaSvc *media.Service) *MediaHandler {
	return &MediaHandler{media: mediaSvc}
}

// Capabilities GET /api/v1/media/capabilities
func (h *MediaHandler) Capabilities(c *gin.Context) {
	cap, err := h.media.Capabilities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cap)
}

// PresignUpload POST /api/v1/media/upload-presign
func (h *MediaHandler) PresignUpload(c *gin.Context) {
	var req struct {
		Kind        string `json:"kind" binding:"required"`
		Ext         string `json:"ext"`
		ContentType string `json:"content_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scope, err := scopeFromKind(req.Kind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if scope == media.ScopeAssistantTTS {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assistant audio is generated server-side only"})
		return
	}
	res, err := h.media.PresignUpload(c.Request.Context(), scope, req.Ext, req.ContentType)
	if err != nil {
		var blocked *media.PresignBlocked
		if errors.As(err, &blocked) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   "PRESIGN_NOT_AVAILABLE",
				"error":  blocked.Reason,
				"reason": blocked.Reason,
			})
			return
		}
		writeMediaStorageError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"method":       res.Method,
		"upload_url":   res.UploadURL,
		"public_url":   res.PublicURL,
		"object_key":   res.ObjectKey,
		"content_type": res.ContentType,
		"expires_at":   res.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

func scopeFromKind(kind string) (media.StorageScope, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "useravatar", "avatar":
		return media.ScopeAvatar, nil
	case "useraudio", "user_recording":
		return media.ScopeUserRecording, nil
	case "aiaudio", "assistant_tts":
		return media.ScopeAssistantTTS, nil
	default:
		return 0, errors.New("unknown kind: use useravatar, useraudio, or aiaudio")
	}
}

func writeMediaStorageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, media.ErrClientOnlyStorage):
		c.JSON(http.StatusForbidden, gin.H{"error": "storage policy is client-only"})
	case errors.Is(err, media.ErrDirectUploadNA):
		c.JSON(http.StatusBadRequest, gin.H{"error": "direct upload not available; use server multipart upload"})
	case errors.Is(err, objectstore.ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "对象存储未配置或凭证不完整，请在管理后台配置并启用"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

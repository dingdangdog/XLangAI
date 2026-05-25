package handler

import (
	"errors"
	"net/http"
	"strings"

	"xlangai/server/internal/authz"
	"xlangai/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type VoiceHandler struct {
	repo *repository.VoiceRepo
}

func NewVoiceHandler(repo *repository.VoiceRepo) *VoiceHandler {
	return &VoiceHandler{repo: repo}
}

func (h *VoiceHandler) List(c *gin.Context) {
	langID := strings.TrimSpace(c.Query("lang_id"))
	if langID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lang_id is required"})
		return
	}

	p := CtxPrincipal(c)
	if p != nil {
		if err := p.EnsureLanguageAllowed(langID); err != nil {
			if errors.Is(err, authz.ErrForbiddenLanguage) {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN_LANGUAGE"})
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
	}

	voices, err := h.repo.ListByLanguage(c.Request.Context(), langID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if p != nil && len(p.Feat.VoiceRoleIDs) > 0 {
		allowed := make(map[string]struct{}, len(p.Feat.VoiceRoleIDs))
		for _, id := range p.Feat.VoiceRoleIDs {
			allowed[id] = struct{}{}
		}
		filtered := make([]*repository.VoiceRole, 0, len(voices))
		for _, v := range voices {
			if _, ok := allowed[v.ID]; ok {
				filtered = append(filtered, v)
			}
		}
		voices = filtered
	}

	c.JSON(http.StatusOK, voices)
}

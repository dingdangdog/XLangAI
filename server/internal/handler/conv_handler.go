package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"xlangai/server/internal/authz"
	"xlangai/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type ConvHandler struct {
	convRepo     *repository.ConvRepo
	msgRepo      *repository.MessageRepo
	systemRepo   *repository.SystemRepo
	scenarioRepo *repository.ScenarioRepo
	userRepo     *repository.UserRepo
	voiceRepo    *repository.VoiceRepo
	langRepo     *repository.LangRepo
	verbose      bool
}

func NewConvHandler(convRepo *repository.ConvRepo, msgRepo *repository.MessageRepo, systemRepo *repository.SystemRepo, scenarioRepo *repository.ScenarioRepo, userRepo *repository.UserRepo, voiceRepo *repository.VoiceRepo, langRepo *repository.LangRepo, verbose bool) *ConvHandler {
	return &ConvHandler{convRepo: convRepo, msgRepo: msgRepo, systemRepo: systemRepo, scenarioRepo: scenarioRepo, userRepo: userRepo, voiceRepo: voiceRepo, langRepo: langRepo, verbose: verbose}
}

func (h *ConvHandler) Create(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		LangID       string  `json:"lang_id" binding:"required"`
		VoiceRoleID  *string `json:"voice_role_id"`
		ScenarioCode *string `json:"scenario_code"`
		Title        *string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := CtxPrincipal(c)
	if p == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := p.EnsureLanguageAllowed(req.LangID); err != nil {
		if errors.Is(err, authz.ErrForbiddenLanguage) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN_LANGUAGE"})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	def, err := h.systemRepo.GetDefaults(c.Request.Context(), req.LangID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get defaults"})
		return
	}
	voiceID := def.VoiceID
	if req.VoiceRoleID != nil && *req.VoiceRoleID != "" {
		voice, err := h.voiceRepo.GetByID(c.Request.Context(), *req.VoiceRoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "voice role not found"})
			return
		}
		if voice.LanguageID != "" && voice.LanguageID != req.LangID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "voice role does not match language"})
			return
		}
		if err := p.EnsureVoiceAllowed(req.VoiceRoleID); err != nil {
			if errors.Is(err, authz.ErrForbiddenVoice) {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN_VOICE"})
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		voiceID = *req.VoiceRoleID
	}

	scenarioCode := "free"
	if req.ScenarioCode != nil {
		sc := strings.TrimSpace(*req.ScenarioCode)
		if sc != "" {
			scenarioCode = sc
		}
	}
	if h.scenarioRepo != nil {
		if _, err := h.scenarioRepo.GetByCode(c.Request.Context(), scenarioCode); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenario_code"})
			return
		}
	}

	promptID, err := h.systemRepo.ResolvePromptIDForScenario(c.Request.Context(), scenarioCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve scenario prompt"})
		return
	}

	titleVal := ""
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t != "" {
			if utf8.RuneCountInString(t) > 200 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "title too long"})
				return
			}
			titleVal = t
		}
	}
	if titleVal == "" && h.scenarioRepo != nil {
		titleVal = h.scenarioRepo.NameByCode(c.Request.Context(), scenarioCode)
	}
	if titleVal == "" {
		titleVal = "New Chat"
	}

	llmConfigID := def.LLMConfigID
	if h.userRepo != nil {
		if userLLM, uerr := h.userRepo.ResolveActiveDefaultLlmConfigID(c.Request.Context(), uid); uerr == nil && userLLM != nil {
			llmConfigID = *userLLM
		}
	}

	conv, err := h.convRepo.Create(c.Request.Context(), uid, req.LangID, voiceID, llmConfigID, promptID, scenarioCode, titleVal)
	if err != nil {
		log.Printf("ConvHandler.Create: user_id=%s lang_id=%s: %v", uid, req.LangID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	enrichConversationScenario(c, h.scenarioRepo, conv)
	c.JSON(http.StatusOK, conv)
}

func (h *ConvHandler) List(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, _ := strconv.Atoi(l); n > 0 && n <= 100 {
			limit = n
		}
	}
	convs, err := h.convRepo.ListByUser(c.Request.Context(), uid, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, conv := range convs {
		enrichConversationScenario(c, h.scenarioRepo, conv)
	}
	c.JSON(http.StatusOK, convs)
}

func (h *ConvHandler) Get(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	conv, err := h.convRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if conv.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if h.langRepo != nil && strings.TrimSpace(conv.LanguageID) != "" {
		if code, err := h.langRepo.GetCodeByID(c.Request.Context(), conv.LanguageID); err == nil {
			conv.LanguageCode = strings.TrimSpace(code)
		}
	}
	enrichConversationScenario(c, h.scenarioRepo, conv)
	c.JSON(http.StatusOK, conv)
}

func (h *ConvHandler) ListMessages(c *gin.Context) {
	id := c.Param("id")
	// 校验会话属于当前用户
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	conv, err := h.convRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if conv.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, _ := strconv.Atoi(l); n > 0 && n <= 100 {
			limit = n
		}
	}
	var beforeID *string
	if b := c.Query("before"); b != "" {
		beforeID = &b
	}
	msgs, err := h.msgRepo.ListByConversation(c.Request.Context(), id, limit, beforeID)
	if err != nil {
		log.Printf("ListMessages: conversation_id=%s limit=%d: %v", id, limit, err)
		if h.verbose {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "failed to list messages",
				"detail": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

func (h *ConvHandler) UpdateVoice(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conv, err := h.convRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if conv.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		VoiceRoleID *string `json:"voice_role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.VoiceRoleID == nil || *req.VoiceRoleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "voice_role_id is required"})
		return
	}

	voice, err := h.voiceRepo.GetByID(c.Request.Context(), *req.VoiceRoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "voice role not found"})
		return
	}
	if voice.LanguageID != "" && voice.LanguageID != conv.LanguageID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "voice role does not match language"})
		return
	}

	p := CtxPrincipal(c)
	if p != nil {
		if err := p.EnsureVoiceAllowed(req.VoiceRoleID); err != nil {
			if errors.Is(err, authz.ErrForbiddenVoice) {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN_VOICE"})
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
	}

	updated, err := h.convRepo.UpdateVoiceRole(c.Request.Context(), id, req.VoiceRoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Patch 更新会话可编辑字段（当前仅支持 title）。
func (h *ConvHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conv, err := h.convRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if conv.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := strings.TrimSpace(req.Title)
	if t == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if utf8.RuneCountInString(t) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title too long"})
		return
	}

	updated, err := h.convRepo.UpdateTitle(c.Request.Context(), id, t)
	if err != nil {
		log.Printf("ConvHandler.Patch: conversation_id=%s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete 软删除当前用户的会话（设置 deleted_at）。
func (h *ConvHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ok, err := h.convRepo.SoftDeleteForUser(c.Request.Context(), id, uid)
	if err != nil {
		log.Printf("ConvHandler.Delete: conversation_id=%s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

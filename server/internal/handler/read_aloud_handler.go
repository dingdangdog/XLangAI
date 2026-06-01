package handler

import (
	"errors"
	"net/http"
	"strings"

	"xlangai/server/internal/authz"
	"xlangai/server/internal/model"
	"xlangai/server/internal/readaloud"
	"xlangai/server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReadAloudHandler struct {
	repo     *repository.ReadAloudRepo
	langRepo *repository.LangRepo
}

func NewReadAloudHandler(repo *repository.ReadAloudRepo, langRepo *repository.LangRepo) *ReadAloudHandler {
	return &ReadAloudHandler{repo: repo, langRepo: langRepo}
}

func (h *ReadAloudHandler) ListCategories(c *gin.Context) {
	categories, err := h.repo.ListActiveCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	langID := strings.TrimSpace(c.Query("lang_id"))
	langCode := ""
	if langID != "" && h.langRepo != nil {
		langCode, _ = h.langRepo.GetCodeByID(c.Request.Context(), langID)
	}
	for _, cat := range categories {
		if cat == nil {
			continue
		}
		if langID != "" {
			n, err := h.repo.CountVocabularies(c.Request.Context(), cat.ID, langID)
			if err == nil {
				cat.VocabCount = n
			}
		}
		if langID != "" {
			if loc, err := h.repo.GetCategoryLocale(c.Request.Context(), cat.ID, langID); err == nil && loc != nil {
				cat.DisplayName = strings.TrimSpace(loc.Name)
				if loc.Description != nil {
					cat.DisplayDescription = strings.TrimSpace(*loc.Description)
				}
			} else if langCode != "" {
				if ent, err := h.repo.GetCategoryByID(c.Request.Context(), cat.ID); err == nil && ent != nil {
					cat.DisplayName, cat.DisplayDescription = readaloud.ResolveCategoryDisplay(ent, langCode)
				}
			}
		}
	}
	c.JSON(http.StatusOK, categories)
}

func (h *ReadAloudHandler) ListVocabularies(c *gin.Context) {
	categoryID := strings.TrimSpace(c.Param("id"))
	langID := strings.TrimSpace(c.Query("lang_id"))
	if categoryID == "" || langID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id and lang_id required"})
		return
	}
	if _, err := h.repo.GetCategoryByID(c.Request.Context(), categoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	items, err := h.repo.ListVocabularies(c.Request.Context(), categoryID, langID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ReadAloudHandler) CreateSession(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		LangID     string `json:"lang_id" binding:"required"`
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
	if _, err := h.repo.GetCategoryByID(c.Request.Context(), req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	vocabs, err := h.repo.ListVocabularies(c.Request.Context(), req.CategoryID, req.LangID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(vocabs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no vocabulary for this category and language"})
		return
	}
	session, err := h.repo.CreateSession(c.Request.Context(), repository.CreateReadAloudSessionInput{
		UserID:     uid,
		CategoryID: req.CategoryID,
		LanguageID: req.LangID,
		TotalItems: len(vocabs),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.enrichSession(c, session)
	c.JSON(http.StatusCreated, gin.H{
		"session":      session,
		"vocabularies": vocabs,
	})
}

func (h *ReadAloudHandler) ListSessions(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sessions, err := h.repo.ListSessions(c.Request.Context(), uid, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range sessions {
		h.enrichSession(c, s)
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *ReadAloudHandler) GetSession(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sessionID := strings.TrimSpace(c.Param("id"))
	row, err := h.repo.GetSession(c.Request.Context(), sessionID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	session := repository.SessionEntityToModel(row)
	h.enrichSession(c, session)
	attempts, err := h.repo.ListAttempts(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"session":  session,
		"attempts": attempts,
	})
}

func (h *ReadAloudHandler) SubmitAttempt(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sessionID := strings.TrimSpace(c.Param("id"))
	if _, err := h.repo.GetSession(c.Request.Context(), sessionID, uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		VocabularyID  string  `json:"vocabulary_id" binding:"required"`
		Part          string  `json:"part" binding:"required"`
		ReferenceText string  `json:"reference_text" binding:"required"`
		Transcript    string  `json:"transcript" binding:"required"`
		Score         int     `json:"score"`
		MatchDetail   *string `json:"match_detail"`
		DurationMs    *int    `json:"duration_ms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	part := strings.TrimSpace(strings.ToLower(req.Part))
	if part != "word" && part != "sentence" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "part must be word or sentence"})
		return
	}
	attempt, err := h.repo.CreateAttempt(c.Request.Context(), repository.CreateReadAloudAttemptInput{
		SessionID:     sessionID,
		VocabularyID:  req.VocabularyID,
		Part:          part,
		ReferenceText: strings.TrimSpace(req.ReferenceText),
		Transcript:    strings.TrimSpace(req.Transcript),
		Score:         req.Score,
		MatchDetail:   req.MatchDetail,
		DurationMs:    req.DurationMs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	session, err := h.repo.RefreshSessionProgress(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.enrichSession(c, session)
	c.JSON(http.StatusCreated, gin.H{
		"attempt": attempt,
		"session": session,
	})
}

func (h *ReadAloudHandler) enrichSession(c *gin.Context, session *model.ReadAloudSession) {
	if session == nil || h.langRepo == nil {
		return
	}
	if cat, err := h.repo.GetCategoryByID(c.Request.Context(), session.CategoryID); err == nil && cat != nil {
		if loc, locErr := h.repo.GetCategoryLocale(
			c.Request.Context(), session.CategoryID, session.LanguageID,
		); locErr == nil && loc != nil {
			session.CategoryName = strings.TrimSpace(loc.Name)
		} else {
			langCode := ""
			if h.langRepo != nil {
				langCode, _ = h.langRepo.GetCodeByID(c.Request.Context(), session.LanguageID)
			}
			session.CategoryName, _ = readaloud.ResolveCategoryDisplay(cat, langCode)
		}
	}
	if lang, err := h.langRepo.GetCodeByID(c.Request.Context(), session.LanguageID); err == nil && lang != "" {
		session.LanguageCode = lang
	}
}

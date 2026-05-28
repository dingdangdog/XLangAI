package handler

import (
	"errors"
	"net/http"
	"strconv"

	"xlangai/server/internal/captcha"

	"github.com/gin-gonic/gin"
)

type CaptchaHandler struct {
	store *captcha.Store
}

func NewCaptchaHandler(store *captcha.Store) *CaptchaHandler {
	return &CaptchaHandler{store: store}
}

// CreateTicket POST /api/v1/captcha/ticket — 获取数学题与 ticket。
func (h *CaptchaHandler) CreateTicket(c *gin.Context) {
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "captcha unavailable", "code": "CAPTCHA_UNAVAILABLE"})
		return
	}
	ticket, question, expiresIn, err := h.store.CreateTicket(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create captcha", "code": "CAPTCHA_CREATE_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ticket":     ticket,
		"question":   question,
		"expires_in": expiresIn,
	})
}

// Verify POST /api/v1/captcha/verify — 提交答案，通过后 ticket 可在 60s 内用于发短信。
func (h *CaptchaHandler) Verify(c *gin.Context) {
	if h.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "captcha unavailable", "code": "CAPTCHA_UNAVAILABLE"})
		return
	}
	var req struct {
		Ticket string      `json:"ticket"`
		Answer interface{} `json:"answer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "INVALID_REQUEST"})
		return
	}
	ans, err := parseCaptchaAnswer(req.Answer)
	if err != nil {
		authAPIError(c, http.StatusBadRequest, "invalid answer", "INVALID_CAPTCHA_ANSWER")
		return
	}
	if err := h.store.Verify(c.Request.Context(), req.Ticket, ans); err != nil {
		mapCaptchaError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func parseCaptchaAnswer(v interface{}) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case string:
		if x == "" {
			return 0, captcha.ErrInvalidAnswer
		}
		return strconv.ParseFloat(x, 64)
	default:
		return 0, captcha.ErrInvalidAnswer
	}
}

func mapCaptchaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, captcha.ErrTicketNotFound):
		authAPIError(c, http.StatusBadRequest, "captcha expired, please refresh", "CAPTCHA_EXPIRED")
	case errors.Is(err, captcha.ErrWrongAnswer), errors.Is(err, captcha.ErrInvalidAnswer):
		authAPIError(c, http.StatusBadRequest, "wrong answer", "INVALID_CAPTCHA_ANSWER")
	case errors.Is(err, captcha.ErrAlreadyConsumed):
		authAPIError(c, http.StatusBadRequest, "captcha expired, please refresh", "CAPTCHA_EXPIRED")
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "CAPTCHA_VERIFY_FAILED"})
	}
}

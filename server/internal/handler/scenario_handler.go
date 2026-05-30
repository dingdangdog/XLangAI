package handler

import (
	"net/http"
	"strings"

	"xlangai/server/internal/model"
	"xlangai/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type ScenarioHandler struct {
	repo *repository.ScenarioRepo
}

func NewScenarioHandler(repo *repository.ScenarioRepo) *ScenarioHandler {
	return &ScenarioHandler{repo: repo}
}

func (h *ScenarioHandler) List(c *gin.Context) {
	scenarios, err := h.repo.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scenarios)
}

func enrichConversationScenario(ctx *gin.Context, repo *repository.ScenarioRepo, conv *model.Conversation) {
	if conv == nil || repo == nil {
		return
	}
	code := conv.ScenarioCode
	if code == "" {
		code = "free"
		conv.ScenarioCode = code
	}
	name := repo.NameByCode(ctx.Request.Context(), code)
	if name == "" && code == "free" {
		name = "自由对话"
	}
	conv.ScenarioName = name
}

func scenarioCodePtr(code string) *string {
	c := strings.TrimSpace(code)
	if c == "" {
		return nil
	}
	return &c
}

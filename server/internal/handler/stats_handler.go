package handler

import (
	"net/http"
	"strconv"
	"time"

	"xlangai/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	stats *repository.StatsRepo
	usage *repository.UsageRepo
}

func NewStatsHandler(stats *repository.StatsRepo, usage *repository.UsageRepo) *StatsHandler {
	return &StatsHandler{stats: stats, usage: usage}
}

// Summary 学习分析页：会话数、消息数、今日/本月对话用量。
func (h *StatsHandler) Summary(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx := c.Request.Context()
	convN, err := h.stats.ConversationCountByUser(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	msgN, err := h.stats.MessageCountByUser(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	today, err := h.usage.TodayUsageCount(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	monthU, err := h.usage.MonthUsageCount(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"conversation_count": convN,
		"message_count":      msgN,
		"usage": gin.H{
			"chat_today":      today,
			"chat_this_month": monthU,
		},
	})
}

// Calendar 分析页对话日历：指定年月的每日对话轮次。
func (h *StatsHandler) Calendar(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	now := time.Now().UTC()
	year, err := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	if err != nil || year < 1970 || year > 9999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	month, err := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}
	days, err := h.usage.DailyCountsInMonth(c.Request.Context(), uid, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"year":  year,
		"month": month,
		"days":  days,
	})
}

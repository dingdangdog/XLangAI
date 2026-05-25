package handler

import (
	"net/http"

	"wlltalk/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type MembershipHandler struct {
	repo *repository.MembershipRepo
}

func NewMembershipHandler(repo *repository.MembershipRepo) *MembershipHandler {
	return &MembershipHandler{repo: repo}
}

func (h *MembershipHandler) ListTiers(c *gin.Context) {
	list, err := h.repo.ListPublic(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

package repository

import (
	"context"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

// LLMServiceConfig 对应表 sys_llm_service_configs（多协议对话）。
type LLMServiceConfig struct {
	ID        string
	Code      string
	Name      string
	Protocol  string
	BaseURL   string
	APIKey    string
	ModelCode string
	Config    string
}

type LLMConfigRepo struct {
	db *gorm.DB
}

func NewLLMConfigRepo(db *gorm.DB) *LLMConfigRepo {
	return &LLMConfigRepo{db: db}
}

func llmFromEntity(c *entity.SysLlmServiceConfig) *LLMServiceConfig {
	if c == nil {
		return nil
	}
	protocol := strings.TrimSpace(c.Protocol)
	if protocol == "" {
		protocol = "openai"
	}
	return &LLMServiceConfig{
		ID:        c.ID,
		Code:      c.Code,
		Name:      c.Name,
		Protocol:  protocol,
		BaseURL:   strVal(c.BaseURL),
		APIKey:    strVal(c.APIKey),
		ModelCode: c.ModelCode,
		Config:    strVal(c.Config),
	}
}

func (r *LLMConfigRepo) GetByID(ctx context.Context, id string) (*LLMServiceConfig, error) {
	var row entity.SysLlmServiceConfig
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return llmFromEntity(&row), nil
}

func (r *LLMConfigRepo) GetByCode(ctx context.Context, code string) (*LLMServiceConfig, error) {
	var row entity.SysLlmServiceConfig
	err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", code, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return llmFromEntity(&row), nil
}

func (r *LLMConfigRepo) GetFirstActive(ctx context.Context) (*LLMServiceConfig, error) {
	var row entity.SysLlmServiceConfig
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return llmFromEntity(&row), nil
}

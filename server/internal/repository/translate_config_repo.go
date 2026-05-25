package repository

import (
	"context"

	"wlltalk/server/internal/entity"

	"gorm.io/gorm"
)

// TranslateServiceConfig 对应表 sys_translate_service_configs。
type TranslateServiceConfig struct {
	ID          string
	Code        string
	Name        string
	Protocol    string
	BaseURL     string
	APIKey      string
	APISecret   string
	ModelCode   string
	LlmConfigID string
	Config      string
}

type TranslateConfigRepo struct {
	db *gorm.DB
}

func NewTranslateConfigRepo(db *gorm.DB) *TranslateConfigRepo {
	return &TranslateConfigRepo{db: db}
}

func translateFromEntity(c *entity.SysTranslateServiceConfig) *TranslateServiceConfig {
	if c == nil {
		return nil
	}
	return &TranslateServiceConfig{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		Protocol:    c.Protocol,
		BaseURL:     strVal(c.BaseURL),
		APIKey:      strVal(c.APIKey),
		APISecret:   strVal(c.APISecret),
		ModelCode:   c.ModelCode,
		LlmConfigID: strVal(c.LlmConfigID),
		Config:      strVal(c.Config),
	}
}

// GetActive 返回唯一启用的翻译配置（按 sort_order、created_at 排序取第一条）。
func (r *TranslateConfigRepo) GetActive(ctx context.Context) (*TranslateServiceConfig, error) {
	var row entity.SysTranslateServiceConfig
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return translateFromEntity(&row), nil
}

func (r *TranslateConfigRepo) GetByID(ctx context.Context, id string) (*TranslateServiceConfig, error) {
	var row entity.SysTranslateServiceConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return translateFromEntity(&row), nil
}

package repository

import (
	"context"

	"wlltalk/server/internal/entity"

	"gorm.io/gorm"
)

// TtsServiceConfig 对应表 sys_tts_service_configs；provider 决定服务端实现（openai_rest / azure_speech_rest）。
type TtsServiceConfig struct {
	ID         string
	Code       string
	Name       string
	Provider   string
	BaseURL    string
	APIKey     string
	Region     string
	ModelCode  string
	ConfigJSON string
}

type TtsConfigRepo struct {
	db *gorm.DB
}

func NewTtsConfigRepo(db *gorm.DB) *TtsConfigRepo {
	return &TtsConfigRepo{db: db}
}

func ttsFromEntity(c *entity.TtsServiceConfig) *TtsServiceConfig {
	if c == nil {
		return nil
	}
	return &TtsServiceConfig{
		ID:         c.ID,
		Code:       c.Code,
		Name:       c.Name,
		Provider:   c.Provider,
		BaseURL:    strVal(c.BaseURL),
		APIKey:     strVal(c.APIKey),
		Region:     strVal(c.Region),
		ModelCode:  c.ModelCode,
		ConfigJSON: strVal(c.Config),
	}
}

func (r *TtsConfigRepo) GetByID(ctx context.Context, id string) (*TtsServiceConfig, error) {
	var row entity.TtsServiceConfig
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return ttsFromEntity(&row), nil
}

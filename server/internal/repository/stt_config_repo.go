package repository

import (
	"context"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

// STTServiceConfig 对应表 sys_stt_service_configs。
// protocol=openai：OpenAI 兼容 POST /v1/audio/transcriptions；protocol=azure_speech_rest：Azure 短音频 REST。
type STTServiceConfig struct {
	ID        string
	Code      string
	Name      string
	Protocol  string
	BaseURL   string
	APIKey    string
	ModelCode string
	Config    string
}

type STTConfigRepo struct {
	db *gorm.DB
}

func NewSTTConfigRepo(db *gorm.DB) *STTConfigRepo {
	return &STTConfigRepo{db: db}
}

func sttFromEntity(c *entity.SysSttServiceConfig) *STTServiceConfig {
	if c == nil {
		return nil
	}
	proto := strings.TrimSpace(c.Protocol)
	if proto == "" {
		proto = "openai"
	}
	return &STTServiceConfig{
		ID:        c.ID,
		Code:      c.Code,
		Name:      c.Name,
		Protocol:  proto,
		BaseURL:   strVal(c.BaseURL),
		APIKey:    strVal(c.APIKey),
		ModelCode: c.ModelCode,
		Config:    strVal(c.Config),
	}
}

func (r *STTConfigRepo) GetFirstActive(ctx context.Context) (*STTServiceConfig, error) {
	var row entity.SysSttServiceConfig
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return sttFromEntity(&row), nil
}

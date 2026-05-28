package repository

import (
	"context"
	"encoding/json"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

// SmsServiceConfig 对应表 sys_sms_service_configs。
type SmsServiceConfig struct {
	ID           string
	Code         string
	Name         string
	Provider     string
	APIKey       string
	SecretKey    string
	Region       string
	SignName     string
	TemplateCode string
	Config       map[string]string
	Status       string
}

type SmsConfigRepo struct {
	db *gorm.DB
}

func NewSmsConfigRepo(db *gorm.DB) *SmsConfigRepo {
	return &SmsConfigRepo{db: db}
}

func smsFromEntity(c *entity.SysSmsServiceConfig) *SmsServiceConfig {
	if c == nil {
		return nil
	}
	out := &SmsServiceConfig{
		ID:       c.ID,
		Code:     c.Code,
		Name:     c.Name,
		Provider: strings.TrimSpace(c.Provider),
		Status:   c.Status,
	}
	if c.APIKey != nil {
		out.APIKey = strings.TrimSpace(*c.APIKey)
	}
	if c.SecretKey != nil {
		out.SecretKey = strings.TrimSpace(*c.SecretKey)
	}
	if c.Region != nil {
		out.Region = strings.TrimSpace(*c.Region)
	}
	if c.SignName != nil {
		out.SignName = strings.TrimSpace(*c.SignName)
	}
	if c.TemplateCode != nil {
		out.TemplateCode = strings.TrimSpace(*c.TemplateCode)
	}
	if c.Config != nil && strings.TrimSpace(*c.Config) != "" {
		_ = json.Unmarshal([]byte(*c.Config), &out.Config)
	}
	if out.Config == nil {
		out.Config = map[string]string{}
	}
	return out
}

func (r *SmsConfigRepo) GetActive(ctx context.Context) (*SmsServiceConfig, error) {
	var row entity.SysSmsServiceConfig
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, updated_at DESC").
		First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return smsFromEntity(&row), nil
}

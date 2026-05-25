package repository

import (
	"context"
	"encoding/json"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

// ObjectStorageConfig 对应表 sys_object_storage_configs。
type ObjectStorageConfig struct {
	ID            string
	Code          string
	Name          string
	Provider      string
	BaseURL       string
	PublicBaseURL string
	APIKey        string
	SecretKey     string
	Bucket        string
	Region        string
	Config        map[string]string
	Status        string
}

type ObjectStorageConfigRepo struct {
	db *gorm.DB
}

func NewObjectStorageConfigRepo(db *gorm.DB) *ObjectStorageConfigRepo {
	return &ObjectStorageConfigRepo{db: db}
}

func objectStorageFromEntity(c *entity.SysObjectStorageConfig) *ObjectStorageConfig {
	if c == nil {
		return nil
	}
	out := &ObjectStorageConfig{
		ID:       c.ID,
		Code:     c.Code,
		Name:     c.Name,
		Provider: strings.TrimSpace(c.Provider),
		Status:   c.Status,
	}
	if c.BaseURL != nil {
		out.BaseURL = strings.TrimSpace(*c.BaseURL)
	}
	if c.PublicBaseURL != nil {
		out.PublicBaseURL = strings.TrimSpace(*c.PublicBaseURL)
	}
	if c.APIKey != nil {
		out.APIKey = strings.TrimSpace(*c.APIKey)
	}
	if c.SecretKey != nil {
		out.SecretKey = strings.TrimSpace(*c.SecretKey)
	}
	if c.Bucket != nil {
		out.Bucket = strings.TrimSpace(*c.Bucket)
	}
	if c.Region != nil {
		out.Region = strings.TrimSpace(*c.Region)
	}
	if c.Config != nil && strings.TrimSpace(*c.Config) != "" {
		_ = json.Unmarshal([]byte(*c.Config), &out.Config)
	}
	if out.Config == nil {
		out.Config = map[string]string{}
	}
	return out
}

func (r *ObjectStorageConfigRepo) GetActive(ctx context.Context) (*ObjectStorageConfig, error) {
	var row entity.SysObjectStorageConfig
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
	return objectStorageFromEntity(&row), nil
}

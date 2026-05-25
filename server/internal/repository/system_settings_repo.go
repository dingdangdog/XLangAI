package repository

import (
	"context"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

type SystemSettingRow struct {
	Key       string
	Value     string
	ValueType string
}

type SystemSettingsRepo struct {
	db *gorm.DB
}

func NewSystemSettingsRepo(db *gorm.DB) *SystemSettingsRepo {
	return &SystemSettingsRepo{db: db}
}

func (r *SystemSettingsRepo) GetByKey(ctx context.Context, key string) (*SystemSettingRow, error) {
	var row entity.SysSystemSetting
	err := r.db.WithContext(ctx).Where(`"key" = ?`, key).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &SystemSettingRow{
		Key:       row.Key,
		Value:     row.Value,
		ValueType: row.ValueType,
	}, nil
}

func (r *SystemSettingsRepo) ListAll(ctx context.Context) ([]SystemSettingRow, error) {
	var rows []entity.SysSystemSetting
	err := r.db.WithContext(ctx).Order(`"key" ASC`).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]SystemSettingRow, len(rows))
	for i := range rows {
		out[i] = SystemSettingRow{
			Key:       rows[i].Key,
			Value:     rows[i].Value,
			ValueType: rows[i].ValueType,
		}
	}
	return out, nil
}

func (r *SystemSettingsRepo) ListByKeys(ctx context.Context, keys []string) ([]SystemSettingRow, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	var rows []entity.SysSystemSetting
	err := r.db.WithContext(ctx).Where(`"key" IN ?`, keys).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]SystemSettingRow, len(rows))
	for i := range rows {
		out[i] = SystemSettingRow{
			Key:       rows[i].Key,
			Value:     rows[i].Value,
			ValueType: rows[i].ValueType,
		}
	}
	return out, nil
}

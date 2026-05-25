package settings

import (
	"context"
	"errors"
	"strings"

	"wlltalk/server/internal/repository"

	"gorm.io/gorm"
)

type Service struct {
	repo *repository.SystemSettingsRepo
}

func NewService(repo *repository.SystemSettingsRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) raw(ctx context.Context, key string) string {
	row, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if v, ok := Defaults[key]; ok {
				return v
			}
			return ""
		}
		return ""
	}
	return strings.TrimSpace(row.Value)
}

func (s *Service) Bool(ctx context.Context, key string) bool {
	v := strings.ToLower(s.raw(ctx, key))
	return v == "true" || v == "1" || v == "yes"
}

func (s *Service) String(ctx context.Context, key string) string {
	return s.raw(ctx, key)
}

// PublicMap 返回可暴露给客户端的 key→value（bool 转为 true/false 字符串保持原样）。
func (s *Service) PublicMap(ctx context.Context) (map[string]interface{}, error) {
	rows, err := s.repo.ListByKeys(ctx, PublicKeys)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]repository.SystemSettingRow, len(rows))
	for _, r := range rows {
		byKey[r.Key] = r
	}
	out := make(map[string]interface{}, len(PublicKeys))
	for _, key := range PublicKeys {
		if row, ok := byKey[key]; ok {
			out[key] = parseValue(row.Value, row.ValueType)
			continue
		}
		def, ok := Defaults[key]
		if !ok {
			continue
		}
		vt := ValueTypeString
		if strings.HasPrefix(key, "auth.") && strings.HasSuffix(key, ".enabled") {
			vt = ValueTypeBool
		}
		if strings.HasSuffix(key, ".register_enabled") {
			vt = ValueTypeBool
		}
		out[key] = parseValue(def, vt)
	}
	return out, nil
}

func parseValue(value, valueType string) interface{} {
	switch valueType {
	case ValueTypeBool:
		v := strings.ToLower(strings.TrimSpace(value))
		return v == "true" || v == "1" || v == "yes"
	default:
		return value
	}
}

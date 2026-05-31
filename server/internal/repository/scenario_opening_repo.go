package repository

import (
	"context"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

type ScenarioOpeningRepo struct {
	db *gorm.DB
}

func NewScenarioOpeningRepo(db *gorm.DB) *ScenarioOpeningRepo {
	return &ScenarioOpeningRepo{db: db}
}

// ResolveTemplate 按场景与语言取开场模板；依次尝试目标语、en、该场景任意一条。
func (r *ScenarioOpeningRepo) ResolveTemplate(ctx context.Context, scenarioCode, langCode string) (string, error) {
	scenarioCode = strings.TrimSpace(scenarioCode)
	langCode = strings.ToLower(strings.TrimSpace(langCode))
	if scenarioCode == "" || scenarioCode == "free" {
		return "", gorm.ErrRecordNotFound
	}

	tryLangs := []string{}
	if langCode != "" {
		tryLangs = append(tryLangs, langCode)
	}
	if langCode != "en" {
		tryLangs = append(tryLangs, "en")
	}
	for _, lc := range tryLangs {
		tpl, err := r.getActive(ctx, scenarioCode, lc)
		if err == nil && strings.TrimSpace(tpl) != "" {
			return tpl, nil
		}
	}

	var row entity.ScenarioOpeningLine
	err := r.db.WithContext(ctx).
		Where("scenario_code = ? AND status = ?", scenarioCode, "active").
		Order("language_code ASC").
		First(&row).Error
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(row.Template), nil
}

func (r *ScenarioOpeningRepo) getActive(ctx context.Context, scenarioCode, langCode string) (string, error) {
	var row entity.ScenarioOpeningLine
	err := r.db.WithContext(ctx).
		Where("scenario_code = ? AND language_code = ? AND status = ?", scenarioCode, langCode, "active").
		First(&row).Error
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(row.Template), nil
}

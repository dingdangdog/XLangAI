package repository

import (
	"context"
	"strings"

	"xlangai/server/internal/entity"
	"xlangai/server/internal/model"

	"gorm.io/gorm"
)

type ScenarioRepo struct {
	db *gorm.DB
}

func NewScenarioRepo(db *gorm.DB) *ScenarioRepo {
	return &ScenarioRepo{db: db}
}

func scenarioToModel(s *entity.PracticeScenario) *model.PracticeScenario {
	if s == nil {
		return nil
	}
	out := &model.PracticeScenario{
		ID:        s.ID,
		Code:      s.Code,
		Name:      s.Name,
		SortOrder: s.SortOrder,
	}
	if s.NameEn != nil {
		out.NameEn = strings.TrimSpace(*s.NameEn)
	}
	if s.Icon != nil {
		out.Icon = strings.TrimSpace(*s.Icon)
	}
	if s.Description != nil {
		out.Description = strings.TrimSpace(*s.Description)
	}
	if s.DescriptionEn != nil {
		out.DescriptionEn = strings.TrimSpace(*s.DescriptionEn)
	}
	return out
}

func (r *ScenarioRepo) ListActive(ctx context.Context) ([]*model.PracticeScenario, error) {
	var rows []entity.PracticeScenario
	if err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.PracticeScenario, 0, len(rows))
	for i := range rows {
		out = append(out, scenarioToModel(&rows[i]))
	}
	return out, nil
}

func (r *ScenarioRepo) GetByCode(ctx context.Context, code string) (*entity.PracticeScenario, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var row entity.PracticeScenario
	err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", code, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ScenarioRepo) NameByCode(ctx context.Context, code string) string {
	row, err := r.GetByCode(ctx, code)
	if err != nil || row == nil {
		return ""
	}
	return strings.TrimSpace(row.Name)
}

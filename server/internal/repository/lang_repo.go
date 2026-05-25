package repository

import (
	"context"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

type Language struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	NameNative string `json:"name_native"`
}

type LangRepo struct {
	db *gorm.DB
}

func NewLangRepo(db *gorm.DB) *LangRepo {
	return &LangRepo{db: db}
}

func (r *LangRepo) GetCodeByID(ctx context.Context, id string) (string, error) {
	var row entity.Language
	err := r.db.WithContext(ctx).
		Select("code").
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	return row.Code, err
}

func (r *LangRepo) List(ctx context.Context) ([]Language, error) {
	var rows []entity.Language
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	list := make([]Language, 0, len(rows))
	for _, row := range rows {
		l := Language{
			ID:   row.ID,
			Code: row.Code,
			Name: row.Name,
		}
		if row.NameNative != nil {
			l.NameNative = *row.NameNative
		}
		list = append(list, l)
	}
	return list, nil
}

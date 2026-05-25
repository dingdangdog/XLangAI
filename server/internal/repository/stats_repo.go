package repository

import (
	"context"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

type StatsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

func (r *StatsRepo) ConversationCountByUser(ctx context.Context, userID string) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&n).Error
	return int(n), err
}

func (r *StatsRepo) MessageCountByUser(ctx context.Context, userID string) (int, error) {
	sub := r.db.Model(&entity.Conversation{}).
		Select("id").
		Where("user_id = ? AND deleted_at IS NULL", userID)
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Message{}).
		Where("conversation_id IN (?)", sub).
		Where("deleted_at IS NULL").
		Count(&n).Error
	return int(n), err
}

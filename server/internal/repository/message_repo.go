package repository

import (
	"context"

	"xlangai/server/internal/entity"
	"xlangai/server/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepo struct {
	db *gorm.DB
}

type CreateMessageInput struct {
	ConversationID   string
	Role             string
	Content          string
	AudioURL         *string
	OriginalAudioURL *string
	STTText          *string
	DurationMs       *int
	Metadata         *string
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) notDeleted() *gorm.DB {
	return r.db.Where("deleted_at IS NULL")
}

func (r *MessageRepo) Create(ctx context.Context, in CreateMessageInput) (*model.Message, error) {
	row := entity.Message{
		ID:               uuid.New().String(),
		ConversationID:   in.ConversationID,
		Role:             in.Role,
		Content:          in.Content,
		AudioURL:         in.AudioURL,
		OriginalAudioURL: in.OriginalAudioURL,
		SttText:          in.STTText,
		DurationMs:       in.DurationMs,
		Metadata:         in.Metadata,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return messageToModel(&row), nil
}

func (r *MessageRepo) CountByConversation(ctx context.Context, convID string) (int64, error) {
	var n int64
	err := r.notDeleted().WithContext(ctx).
		Model(&entity.Message{}).
		Where("conversation_id = ?", convID).
		Count(&n).Error
	return n, err
}

func (r *MessageRepo) ListByConversation(ctx context.Context, convID string, limit int, beforeID *string) ([]*model.Message, error) {
	q := r.notDeleted().WithContext(ctx).Where("conversation_id = ?", convID)
	if beforeID != nil && *beforeID != "" {
		var anchor entity.Message
		if err := r.notDeleted().WithContext(ctx).
			Where("id = ? AND conversation_id = ?", *beforeID, convID).
			First(&anchor).Error; err != nil {
			return nil, err
		}
		q = q.Where("created_at < ?", anchor.CreatedAt)
	}
	var rows []entity.Message
	if err := q.Order("created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	// 反转为时间正序
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	list := make([]*model.Message, 0, len(rows))
	for i := range rows {
		list = append(list, messageToModel(&rows[i]))
	}
	return list, nil
}

func (r *MessageRepo) UpdateMetadata(ctx context.Context, id string, metadata *string) (*model.Message, error) {
	if err := r.notDeleted().WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ?", id).
		Update("metadata", metadata).Error; err != nil {
		return nil, err
	}
	var row entity.Message
	if err := r.notDeleted().WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return messageToModel(&row), nil
}

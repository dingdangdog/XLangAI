package repository

import (
	"context"
	"strings"
	"time"

	"xlangai/server/internal/entity"
	"xlangai/server/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConvRepo struct {
	db *gorm.DB
}

func NewConvRepo(db *gorm.DB) *ConvRepo {
	return &ConvRepo{db: db}
}

func (r *ConvRepo) notDeleted() *gorm.DB {
	return r.db.Where("deleted_at IS NULL")
}

func (r *ConvRepo) Create(ctx context.Context, userID, langID, voiceRoleID, llmConfigID, promptID, title string) (*model.Conversation, error) {
	t := strings.TrimSpace(title)
	if t == "" {
		t = "New Chat"
	}
	row := entity.Conversation{
		ID:          uuid.New().String(),
		UserID:      userID,
		LanguageID:  langID,
		VoiceRoleID: strPtr(voiceRoleID),
		LlmConfigID: strPtr(llmConfigID),
		PromptID:    strPtr(promptID),
		Title:       &t,
		Status:      "active",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return convToModel(&row), nil
}

func (r *ConvRepo) UpdateTitle(ctx context.Context, id string, title string) (*model.Conversation, error) {
	t := strings.TrimSpace(title)
	var row entity.Conversation
	err := r.notDeleted().WithContext(ctx).Where("id = ?", id).
		Updates(map[string]any{"title": t}).Error
	if err != nil {
		return nil, err
	}
	if err := r.notDeleted().WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return convToModel(&row), nil
}

type convLastMsg struct {
	ConversationID string    `gorm:"column:conversation_id"`
	LastAt         time.Time `gorm:"column:last_at"`
}

func (r *ConvRepo) ListByUser(ctx context.Context, userID string, limit int) ([]*model.Conversation, error) {
	var convs []entity.Conversation
	if err := r.notDeleted().WithContext(ctx).Where("user_id = ?", userID).Find(&convs).Error; err != nil {
		return nil, err
	}
	if len(convs) == 0 {
		return []*model.Conversation{}, nil
	}
	ids := make([]string, len(convs))
	for i := range convs {
		ids[i] = convs[i].ID
	}
	var lasts []convLastMsg
	if err := r.db.WithContext(ctx).Model(&entity.Message{}).
		Select("conversation_id, MAX(created_at) AS last_at").
		Where("conversation_id IN ? AND deleted_at IS NULL", ids).
		Group("conversation_id").
		Scan(&lasts).Error; err != nil {
		return nil, err
	}
	lastMap := make(map[string]time.Time, len(lasts))
	for _, lm := range lasts {
		lastMap[lm.ConversationID] = lm.LastAt
	}
	type sortItem struct {
		conv entity.Conversation
		last time.Time
	}
	items := make([]sortItem, len(convs))
	for i, c := range convs {
		last := c.UpdatedAt
		if t, ok := lastMap[c.ID]; ok && t.After(last) {
			last = t
		}
		items[i] = sortItem{conv: c, last: last}
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].last.After(items[i].last) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if limit > len(items) {
		limit = len(items)
	}
	list := make([]*model.Conversation, 0, limit)
	for i := 0; i < limit; i++ {
		c := convToModel(&items[i].conv)
		la := items[i].last
		c.LastActivityAt = &la
		list = append(list, c)
	}
	return list, nil
}

func (r *ConvRepo) GetByID(ctx context.Context, id string) (*model.Conversation, error) {
	var row entity.Conversation
	err := r.notDeleted().WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return convToModel(&row), nil
}

func (r *ConvRepo) UpdateVoiceRole(ctx context.Context, id string, voiceRoleID *string) (*model.Conversation, error) {
	err := r.notDeleted().WithContext(ctx).Where("id = ?", id).
		Updates(map[string]any{"voice_role_id": voiceRoleID}).Error
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

// SoftDeleteForUser 将会话标记为已删除（仅当 id 属于 userID 且尚未删除时）。返回是否更新到一行。
func (r *ConvRepo) SoftDeleteForUser(ctx context.Context, id, userID string) (bool, error) {
	now := time.Now()
	res := r.notDeleted().WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

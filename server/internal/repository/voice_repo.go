package repository

import (
	"context"
	"strings"

	"wlltalk/server/internal/entity"

	"gorm.io/gorm"
)

type VoiceRole struct {
	ID                 string  `json:"id"`
	LanguageID         string  `json:"language_id"`
	TtsServiceConfigID string  `json:"tts_service_config_id"`
	VoiceCode          string  `json:"voice_code"`
	Name               string  `json:"name"`
	Gender             string  `json:"gender,omitempty"`
	PreviewAudioURL    string  `json:"preview_audio_url,omitempty"`
}

type VoiceRepo struct {
	db *gorm.DB
}

func NewVoiceRepo(db *gorm.DB) *VoiceRepo {
	return &VoiceRepo{db: db}
}

func voiceFromEntity(v *entity.VoiceRole) *VoiceRole {
	if v == nil {
		return nil
	}
	out := &VoiceRole{
		ID:        v.ID,
		VoiceCode: v.VoiceCode,
		Name:      v.Name,
	}
	if v.LanguageID != nil {
		out.LanguageID = *v.LanguageID
	}
	if v.TtsServiceConfigID != nil {
		out.TtsServiceConfigID = *v.TtsServiceConfigID
	}
	if v.Gender != nil {
		out.Gender = *v.Gender
	}
	if v.PreviewAudioURL != nil {
		out.PreviewAudioURL = strings.TrimSpace(*v.PreviewAudioURL)
	}
	return out
}

func (r *VoiceRepo) GetByID(ctx context.Context, id string) (*VoiceRole, error) {
	var row entity.VoiceRole
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return voiceFromEntity(&row), nil
}

func (r *VoiceRepo) ListByLanguage(ctx context.Context, languageID string) ([]*VoiceRole, error) {
	var rows []entity.VoiceRole
	err := r.db.WithContext(ctx).
		Where("language_id = ? AND status = ?", languageID, "active").
		Order("sort_order ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	list := make([]*VoiceRole, 0, len(rows))
	for i := range rows {
		list = append(list, voiceFromEntity(&rows[i]))
	}
	return list, nil
}

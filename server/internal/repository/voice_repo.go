package repository

import (
	"context"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

type VoiceRole struct {
	ID                 string `json:"id"`
	LanguageID         string `json:"language_id"`
	SynthesisType      string `json:"synthesis_type"`
	LlmServiceConfigID string `json:"llm_service_config_id,omitempty"`
	TtsServiceConfigID string `json:"tts_service_config_id"`
	VoiceCode          string `json:"voice_code"`
	Name               string `json:"name"`
	Gender             string `json:"gender,omitempty"`
	PreviewAudioURL    string `json:"preview_audio_url,omitempty"`
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
		ID:            v.ID,
		VoiceCode:     v.VoiceCode,
		Name:          v.Name,
		SynthesisType: NormalizeSynthesisType(v.SynthesisType),
	}
	if v.LanguageID != nil {
		out.LanguageID = *v.LanguageID
	}
	if v.LlmServiceConfigID != nil {
		out.LlmServiceConfigID = strings.TrimSpace(*v.LlmServiceConfigID)
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

func (r *VoiceRepo) loadRow(ctx context.Context, id string) (*entity.VoiceRole, error) {
	var row entity.VoiceRole
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	if row.SynthesisType == "" {
		row.SynthesisType = SynthesisTTS
	}
	return &row, nil
}

func (r *VoiceRepo) GetByID(ctx context.Context, id string) (*VoiceRole, error) {
	row, err := r.loadRow(ctx, id)
	if err != nil {
		return nil, err
	}
	return voiceFromEntity(row), nil
}

// GetEntityByID 返回完整实体（含 rolePrompt、config），供对话编排使用。
func (r *VoiceRepo) GetEntityByID(ctx context.Context, id string) (*entity.VoiceRole, error) {
	return r.loadRow(ctx, id)
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
		if rows[i].SynthesisType == "" {
			rows[i].SynthesisType = SynthesisTTS
		}
		list = append(list, voiceFromEntity(&rows[i]))
	}
	return list, nil
}

package repository

import (
	"strings"

	"xlangai/server/internal/entity"
	"xlangai/server/internal/messagemeta"
	"xlangai/server/internal/model"
)

func userToModel(u *entity.User) *model.User {
	if u == nil {
		return nil
	}
	out := &model.User{
		ID:                    u.ID,
		Phone:                 u.Phone,
		Email:                 u.Email,
		Nickname:              u.Nickname,
		AvatarURL:             u.AvatarURL,
		TierID:                u.TierID,
		LanguageID:            u.LanguageID,
		TokenBalance:          u.TokenBalance,
		SubscriptionExpiresAt: u.SubscriptionExpiresAt,
		Status:                u.Status,
		LastLoginAt:           u.LastLoginAt,
		CreatedAt:             u.CreatedAt,
		UpdatedAt:             u.UpdatedAt,
	}
	if u.PasswordHash != nil {
		out.PasswordHash = *u.PasswordHash
	}
	return out
}

func convToModel(c *entity.Conversation) *model.Conversation {
	if c == nil {
		return nil
	}
	title := "新对话"
	if c.Title != nil && strings.TrimSpace(*c.Title) != "" {
		title = strings.TrimSpace(*c.Title)
	}
	return &model.Conversation{
		ID:           c.ID,
		UserID:       c.UserID,
		LanguageID:   c.LanguageID,
		VoiceRoleID:  c.VoiceRoleID,
		LLMConfigID:  c.LlmConfigID,
		PromptID:     c.PromptID,
		ScenarioCode: strVal(c.ScenarioCode),
		Title:        title,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func messageToModel(m *entity.Message) *model.Message {
	if m == nil {
		return nil
	}
	status := messagemeta.ParseStatus(m.Metadata)
	return &model.Message{
		ID:                  m.ID,
		ConversationID:      m.ConversationID,
		Role:                m.Role,
		Content:             m.Content,
		AudioURL:            m.AudioURL,
		OriginalAudioURL:    m.OriginalAudioURL,
		STTText:             m.SttText,
		DurationMs:          m.DurationMs,
		Metadata:            m.Metadata,
		AIInteractionStatus: status,
		CreatedAt:           m.CreatedAt,
	}
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

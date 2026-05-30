package model

import "time"

type User struct {
	ID           string     `json:"id"`
	Phone        *string    `json:"phone,omitempty"`
	Email        *string    `json:"email,omitempty"`
	PasswordHash string     `json:"-"`
	Nickname     *string    `json:"nickname,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	Role         string     `json:"role,omitempty"`
	TierID       *string    `json:"tier_id,omitempty"`
	LanguageID   *string    `json:"language_id,omitempty"`
	TokenBalance int64      `json:"token_balance"`
	// SubscriptionExpiresAt 商店订阅权益到期（UTC）；nil 表示未记录或非订阅来源。
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	Status                  string     `json:"status"`
	LastLoginAt             *time.Time `json:"last_login_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type Conversation struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	LanguageID     string     `json:"language_id"`
	LanguageCode   string     `json:"language_code,omitempty"`
	VoiceRoleID    *string    `json:"voice_role_id,omitempty"`
	LLMConfigID    *string    `json:"ai_config_id,omitempty"`
	PromptID       *string    `json:"prompt_id,omitempty"`
	ScenarioCode   string     `json:"scenario_code,omitempty"`
	ScenarioName   string     `json:"scenario_name,omitempty"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

type PracticeScenario struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	NameEn      string  `json:"name_en,omitempty"`
	Icon        string  `json:"icon,omitempty"`
	Description string  `json:"description,omitempty"`
	DescriptionEn string  `json:"description_en,omitempty"`
	SortOrder   int     `json:"sort_order"`
}

type Message struct {
	ID                  string    `json:"id"`
	ConversationID      string    `json:"conversation_id"`
	Role                string    `json:"role"` // user, assistant, system
	Content             string    `json:"content"`
	AudioURL            *string   `json:"audio_url,omitempty"`
	OriginalAudioURL    *string   `json:"original_audio_url,omitempty"`
	STTText             *string   `json:"stt_text,omitempty"`
	DurationMs          *int      `json:"duration_ms,omitempty"`
	Metadata            *string   `json:"metadata,omitempty"`
	AIInteractionStatus string    `json:"ai_interaction_status,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

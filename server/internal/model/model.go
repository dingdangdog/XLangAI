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

type ReadAloudCategory struct {
	ID              string `json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	NameEn          string `json:"name_en,omitempty"`
	Icon            string `json:"icon,omitempty"`
	Description     string `json:"description,omitempty"`
	DescriptionEn   string `json:"description_en,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	DisplayDescription string `json:"display_description,omitempty"`
	SortOrder       int    `json:"sort_order"`
	VocabCount      int    `json:"vocab_count,omitempty"`
}

type ReadAloudVocabulary struct {
	ID               string     `json:"id"`
	CategoryID       string     `json:"category_id"`
	LanguageID       string     `json:"language_id"`
	Word             string     `json:"word"`
	ExampleSentence  string     `json:"example_sentence"`
	VoiceRoleID      string     `json:"voice_role_id"`
	WordAudioURL     *string    `json:"word_audio_url,omitempty"`
	SentenceAudioURL *string    `json:"sentence_audio_url,omitempty"`
	SortOrder        int        `json:"sort_order"`
}

type ReadAloudSession struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	CategoryID     string     `json:"category_id"`
	CategoryName   string     `json:"category_name,omitempty"`
	LanguageID     string     `json:"language_id"`
	LanguageCode   string     `json:"language_code,omitempty"`
	Status         string     `json:"status"`
	TotalItems     int        `json:"total_items"`
	CompletedItems int        `json:"completed_items"`
	AverageScore   *int       `json:"average_score,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ReadAloudAttempt struct {
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	VocabularyID  string    `json:"vocabulary_id"`
	Part          string    `json:"part"`
	ReferenceText string    `json:"reference_text"`
	Transcript    string    `json:"transcript"`
	Score         int       `json:"score"`
	MatchDetail   *string   `json:"match_detail,omitempty"`
	DurationMs    *int      `json:"duration_ms,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
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

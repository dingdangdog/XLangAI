// Package entity 定义与 manager/prisma/schema.prisma 一致的 GORM 表实体（约定式外键，无 relation）。
package entity

import (
	"time"
)

// --- sys_* 系统配置 ---

type Language struct {
	ID                string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code              string    `gorm:"column:code;type:varchar(10);uniqueIndex"`
	Name              string    `gorm:"column:name;type:varchar(100)"`
	NameNative        *string   `gorm:"column:name_native;type:varchar(100)"`
	PreviewSampleText *string   `gorm:"column:preview_sample_text;type:varchar(500)"`
	SortOrder         int       `gorm:"column:sort_order"`
	Status            string    `gorm:"column:status;type:varchar(20)"`
	Remark            *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (Language) TableName() string { return "sys_languages" }

type SysLlmServiceConfig struct {
	ID        string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code      string    `gorm:"column:code;type:varchar(100);uniqueIndex"`
	Name      string    `gorm:"column:name;type:varchar(200)"`
	Protocol  string    `gorm:"column:protocol;type:varchar(50)"`
	BaseURL   *string   `gorm:"column:base_url;type:varchar(512)"`
	APIKey    *string   `gorm:"column:api_key;type:varchar(500)"`
	ModelCode string    `gorm:"column:model_code;type:varchar(100)"`
	Config    *string   `gorm:"column:config;type:text"`
	Status    string    `gorm:"column:status;type:varchar(20)"`
	SortOrder int       `gorm:"column:sort_order"`
	Remark    *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SysLlmServiceConfig) TableName() string { return "sys_llm_service_configs" }

type SysSttServiceConfig struct {
	ID        string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code      string    `gorm:"column:code;type:varchar(100);uniqueIndex"`
	Name      string    `gorm:"column:name;type:varchar(200)"`
	Protocol  string    `gorm:"column:protocol;type:varchar(50)"`
	BaseURL   *string   `gorm:"column:base_url;type:varchar(512)"`
	APIKey    *string   `gorm:"column:api_key;type:varchar(500)"`
	ModelCode string    `gorm:"column:model_code;type:varchar(100)"`
	Config    *string   `gorm:"column:config;type:text"`
	Status    string    `gorm:"column:status;type:varchar(20)"`
	SortOrder int       `gorm:"column:sort_order"`
	Remark    *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SysSttServiceConfig) TableName() string { return "sys_stt_service_configs" }

type SysTranslateServiceConfig struct {
	ID          string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code        string    `gorm:"column:code;type:varchar(100);uniqueIndex"`
	Name        string    `gorm:"column:name;type:varchar(200)"`
	Protocol    string    `gorm:"column:protocol;type:varchar(50)"`
	BaseURL     *string   `gorm:"column:base_url;type:varchar(512)"`
	APIKey      *string   `gorm:"column:api_key;type:varchar(500)"`
	APISecret   *string   `gorm:"column:api_secret;type:varchar(500)"`
	ModelCode   string    `gorm:"column:model_code;type:varchar(100)"`
	LlmConfigID *string   `gorm:"column:llm_config_id;type:varchar(36)"`
	Config      *string   `gorm:"column:config;type:text"`
	Status      string    `gorm:"column:status;type:varchar(20)"`
	SortOrder   int       `gorm:"column:sort_order"`
	Remark      *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (SysTranslateServiceConfig) TableName() string { return "sys_translate_service_configs" }

type SysObjectStorageConfig struct {
	ID            string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code          string    `gorm:"column:code;type:varchar(100);uniqueIndex"`
	Name          string    `gorm:"column:name;type:varchar(200)"`
	Provider      string    `gorm:"column:provider;type:varchar(50)"`
	BaseURL       *string   `gorm:"column:base_url;type:varchar(512)"`
	PublicBaseURL *string   `gorm:"column:public_base_url;type:varchar(512)"`
	APIKey        *string   `gorm:"column:api_key;type:varchar(500)"`
	SecretKey     *string   `gorm:"column:secret_key;type:varchar(500)"`
	Bucket        *string   `gorm:"column:bucket;type:varchar(128)"`
	Region        *string   `gorm:"column:region;type:varchar(64)"`
	Config        *string   `gorm:"column:config;type:text"`
	Status        string    `gorm:"column:status;type:varchar(20)"`
	SortOrder     int       `gorm:"column:sort_order"`
	Remark        *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (SysObjectStorageConfig) TableName() string { return "sys_object_storage_configs" }

type SysSmsServiceConfig struct {
	ID           string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code         string    `gorm:"column:code;type:varchar(100);uniqueIndex"`
	Name         string    `gorm:"column:name;type:varchar(200)"`
	Provider     string    `gorm:"column:provider;type:varchar(50)"`
	APIKey       *string   `gorm:"column:api_key;type:varchar(500)"`
	SecretKey    *string   `gorm:"column:secret_key;type:varchar(500)"`
	Region       *string   `gorm:"column:region;type:varchar(64)"`
	SignName     *string   `gorm:"column:sign_name;type:varchar(100)"`
	TemplateCode *string   `gorm:"column:template_code;type:varchar(100)"`
	Config       *string   `gorm:"column:config;type:text"`
	Status       string    `gorm:"column:status;type:varchar(20)"`
	SortOrder    int       `gorm:"column:sort_order"`
	Remark       *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (SysSmsServiceConfig) TableName() string { return "sys_sms_service_configs" }

type SysSystemSetting struct {
	ID          string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Key         string    `gorm:"column:key;type:varchar(128);uniqueIndex"`
	Value       string    `gorm:"column:value;type:text"`
	ValueType   string    `gorm:"column:value_type;type:varchar(20)"`
	Status      string    `gorm:"column:status;type:varchar(20)"`
	Description *string   `gorm:"column:description;type:varchar(500)"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (SysSystemSetting) TableName() string { return "sys_system_settings" }

type TtsServiceConfig struct {
	ID        string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code      string    `gorm:"column:code;type:varchar(100);uniqueIndex"`
	Name      string    `gorm:"column:name;type:varchar(200)"`
	Provider  string    `gorm:"column:provider;type:varchar(50)"`
	BaseURL   *string   `gorm:"column:base_url;type:varchar(512)"`
	APIKey    *string   `gorm:"column:api_key;type:varchar(500)"`
	Region    *string   `gorm:"column:region;type:varchar(64)"`
	ModelCode string    `gorm:"column:model_code;type:varchar(100)"`
	Config    *string   `gorm:"column:config;type:text"`
	Status    string    `gorm:"column:status;type:varchar(20)"`
	SortOrder int       `gorm:"column:sort_order"`
	Remark    *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (TtsServiceConfig) TableName() string { return "sys_tts_service_configs" }

type VoiceRole struct {
	ID                   string     `gorm:"column:id;type:varchar(36);primaryKey"`
	LanguageID           *string    `gorm:"column:language_id;type:varchar(36)"`
	SynthesisType        string     `gorm:"column:synthesis_type;type:varchar(30)"`
	LlmServiceConfigID   *string    `gorm:"column:llm_service_config_id;type:varchar(36)"`
	TtsServiceConfigID   *string    `gorm:"column:tts_service_config_id;type:varchar(36)"`
	VoiceCode            string     `gorm:"column:voice_code;type:varchar(50)"`
	Name                 string     `gorm:"column:name;type:varchar(100)"`
	Gender               *string    `gorm:"column:gender;type:varchar(20)"`
	RolePrompt           *string    `gorm:"column:role_prompt;type:text"`
	Config               *string    `gorm:"column:config;type:text"`
	PreviewAudioURL      *string    `gorm:"column:preview_audio_url;type:varchar(500)"`
	PreviewLocalFilename *string    `gorm:"column:preview_local_filename;type:varchar(200)"`
	PreviewGeneratedAt   *time.Time `gorm:"column:preview_generated_at"`
	Status               string     `gorm:"column:status;type:varchar(20)"`
	SortOrder            int        `gorm:"column:sort_order"`
	Remark               *string    `gorm:"column:remark;type:varchar(500)"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (VoiceRole) TableName() string { return "sys_voice_roles" }

type PromptTemplate struct {
	ID         string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code       string    `gorm:"column:code;type:varchar(50);uniqueIndex"`
	Name       string    `gorm:"column:name;type:varchar(200)"`
	Content    string    `gorm:"column:content;type:text"`
	Variables  *string   `gorm:"column:variables;type:text"`
	LanguageID *string   `gorm:"column:language_id;type:varchar(36)"`
	Status     string    `gorm:"column:status;type:varchar(20)"`
	SortOrder  int       `gorm:"column:sort_order"`
	Remark     *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (PromptTemplate) TableName() string { return "sys_prompt_templates" }

type PracticeScenario struct {
	ID               string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code             string    `gorm:"column:code;type:varchar(32);uniqueIndex"`
	Name             string    `gorm:"column:name;type:varchar(100)"`
	NameEn           *string   `gorm:"column:name_en;type:varchar(100)"`
	Icon             *string   `gorm:"column:icon;type:varchar(50)"`
	Description      *string   `gorm:"column:description;type:varchar(500)"`
	DescriptionEn    *string   `gorm:"column:description_en;type:varchar(500)"`
	PromptTemplateID *string   `gorm:"column:prompt_template_id;type:varchar(36)"`
	SortOrder        int       `gorm:"column:sort_order"`
	Status           string    `gorm:"column:status;type:varchar(20)"`
	Remark           *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (PracticeScenario) TableName() string { return "sys_practice_scenarios" }

type ScenarioOpeningLine struct {
	ID           string    `gorm:"column:id;type:varchar(36);primaryKey"`
	ScenarioCode string    `gorm:"column:scenario_code;type:varchar(32)"`
	LanguageCode string    `gorm:"column:language_code;type:varchar(10)"`
	Template     string    `gorm:"column:template;type:varchar(500)"`
	Status       string    `gorm:"column:status;type:varchar(20)"`
	Remark       *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ScenarioOpeningLine) TableName() string { return "sys_scenario_opening_lines" }

type MembershipTier struct {
	ID           string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code         string    `gorm:"column:code;type:varchar(20);uniqueIndex"`
	Name         string    `gorm:"column:name;type:varchar(100)"`
	DailyLimit   *int      `gorm:"column:daily_limit"`
	MonthlyLimit *int      `gorm:"column:monthly_limit"`
	Features     *string   `gorm:"column:features;type:text"`
	Status       string    `gorm:"column:status;type:varchar(20)"`
	SortOrder    int       `gorm:"column:sort_order"`
	Remark       *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (MembershipTier) TableName() string { return "sys_membership_tiers" }

type BillingProduct struct {
	ID               string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code             string    `gorm:"column:code;type:varchar(64);uniqueIndex"`
	Kind             string    `gorm:"column:kind;type:varchar(32)"`
	IOSProductID     string    `gorm:"column:ios_product_id;type:varchar(200)"`
	AndroidProductID string    `gorm:"column:android_product_id;type:varchar(200)"`
	TierCode         *string   `gorm:"column:tier_code;type:varchar(20)"`
	TokenGrant       int64     `gorm:"column:token_grant"`
	SortOrder        int       `gorm:"column:sort_order"`
	Status           string    `gorm:"column:status;type:varchar(20)"`
	Remark           *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (BillingProduct) TableName() string { return "sys_billing_products" }

// --- usr_* 用户域 ---

type User struct {
	ID                    string     `gorm:"column:id;type:varchar(36);primaryKey"`
	Phone                 *string    `gorm:"column:phone;type:varchar(20)"`
	Email                 *string    `gorm:"column:email;type:varchar(255)"`
	PasswordHash          *string    `gorm:"column:password_hash;type:varchar(255)"`
	Nickname              *string    `gorm:"column:nickname;type:varchar(100)"`
	AvatarURL             *string    `gorm:"column:avatar_url;type:varchar(500)"`
	TierID                *string    `gorm:"column:tier_id;type:varchar(36)"`
	LanguageID            *string    `gorm:"column:language_id;type:varchar(36)"`
	DefaultLlmConfigID    *string    `gorm:"column:default_llm_config_id;type:varchar(36)"`
	Settings              *string    `gorm:"column:settings;type:text"`
	TokenBalance          int64      `gorm:"column:token_balance"`
	TurnBalance           int        `gorm:"column:turn_balance"`
	SubscriptionExpiresAt *time.Time `gorm:"column:subscription_expires_at"`
	AppleSub              *string    `gorm:"column:apple_sub;type:varchar(255);uniqueIndex"`
	GoogleSub             *string    `gorm:"column:google_sub;type:varchar(255);uniqueIndex"`
	Status                string     `gorm:"column:status;type:varchar(20)"`
	LastLoginAt           *time.Time `gorm:"column:last_login_at"`
	Remark                *string    `gorm:"column:remark;type:varchar(500)"`
	DeletedAt             *time.Time `gorm:"column:deleted_at"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "usr_users" }

type StoreTransaction struct {
	ID                 string    `gorm:"column:id;type:varchar(36);primaryKey"`
	UserID             string    `gorm:"column:user_id;type:varchar(36)"`
	Platform           string    `gorm:"column:platform;type:varchar(16)"`
	StoreTransactionID string    `gorm:"column:store_transaction_id;type:varchar(512)"`
	ProductCode        string    `gorm:"column:product_code;type:varchar(64)"`
	RawPayload         *string   `gorm:"column:raw_payload;type:text"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (StoreTransaction) TableName() string { return "usr_store_transactions" }

type UserUsage struct {
	ID             string    `gorm:"column:id;type:varchar(36);primaryKey"`
	UserID         string    `gorm:"column:user_id;type:varchar(36)"`
	Date           time.Time `gorm:"column:date;type:date"`
	UsageCount     int       `gorm:"column:usage_count"`
	TokenCount     int       `gorm:"column:token_count"`
	TranslateCount int       `gorm:"column:translate_count"`
	TranslateChars int       `gorm:"column:translate_chars"`
	TtsCount       int       `gorm:"column:tts_count"`
	TtsChars       int       `gorm:"column:tts_chars"`
	SttCount       int       `gorm:"column:stt_count"`
	SttAudioBytes  int64     `gorm:"column:stt_audio_bytes"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (UserUsage) TableName() string { return "usr_user_usage" }

type ServiceUsageDaily struct {
	ID           string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Date         time.Time `gorm:"column:date;type:date"`
	ServiceType  string    `gorm:"column:service_type;type:varchar(20)"`
	ConfigID     string    `gorm:"column:config_id;type:varchar(36)"`
	RequestCount int       `gorm:"column:request_count"`
	UnitCount    int64     `gorm:"column:unit_count"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ServiceUsageDaily) TableName() string { return "sys_service_usage_daily" }

type Conversation struct {
	ID           string     `gorm:"column:id;type:varchar(36);primaryKey"`
	UserID       string     `gorm:"column:user_id;type:varchar(36)"`
	LanguageID   string     `gorm:"column:language_id;type:varchar(36)"`
	VoiceRoleID  *string    `gorm:"column:voice_role_id;type:varchar(36)"`
	LlmConfigID  *string    `gorm:"column:llm_config_id;type:varchar(36)"`
	PromptID     *string    `gorm:"column:prompt_id;type:varchar(36)"`
	ScenarioCode *string    `gorm:"column:scenario_code;type:varchar(32)"`
	Title        *string    `gorm:"column:title;type:varchar(200)"`
	Status       string     `gorm:"column:status;type:varchar(20)"`
	Remark       *string    `gorm:"column:remark;type:varchar(500)"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (Conversation) TableName() string { return "usr_conversations" }

type Message struct {
	ID               string     `gorm:"column:id;type:varchar(36);primaryKey"`
	ConversationID   string     `gorm:"column:conversation_id;type:varchar(36)"`
	Role             string     `gorm:"column:role;type:varchar(20)"`
	Content          string     `gorm:"column:content;type:text"`
	AudioURL         *string    `gorm:"column:audio_url;type:varchar(500)"`
	OriginalAudioURL *string    `gorm:"column:original_audio_url;type:varchar(500)"`
	SttText          *string    `gorm:"column:stt_text;type:text"`
	DurationMs       *int       `gorm:"column:duration_ms"`
	Metadata         *string    `gorm:"column:metadata;type:text"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (Message) TableName() string { return "usr_messages" }

type ReadAloudCategory struct {
	ID             string    `gorm:"column:id;type:varchar(36);primaryKey"`
	Code           string    `gorm:"column:code;type:varchar(32);uniqueIndex"`
	Name           string    `gorm:"column:name;type:varchar(100)"`
	NameEn         *string   `gorm:"column:name_en;type:varchar(100)"`
	Icon           *string   `gorm:"column:icon;type:varchar(50)"`
	Description   *string   `gorm:"column:description;type:varchar(500)"`
	DescriptionEn *string   `gorm:"column:description_en;type:varchar(500)"`
	SortOrder     int       `gorm:"column:sort_order"`
	Status        string    `gorm:"column:status;type:varchar(20)"`
	Remark        *string   `gorm:"column:remark;type:varchar(500)"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (ReadAloudCategory) TableName() string { return "sys_read_aloud_categories" }

type ReadAloudCategoryLocale struct {
	ID          string    `gorm:"column:id;type:varchar(36);primaryKey"`
	CategoryID  string    `gorm:"column:category_id;type:varchar(36)"`
	LanguageID  string    `gorm:"column:language_id;type:varchar(36)"`
	Name        string    `gorm:"column:name;type:varchar(100)"`
	Description *string   `gorm:"column:description;type:varchar(500)"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (ReadAloudCategoryLocale) TableName() string { return "sys_read_aloud_category_locales" }

type ReadAloudVocabulary struct {
	ID                         string     `gorm:"column:id;type:varchar(36);primaryKey"`
	CategoryID                 string     `gorm:"column:category_id;type:varchar(36)"`
	LanguageID                 string     `gorm:"column:language_id;type:varchar(36)"`
	Word                       string     `gorm:"column:word;type:varchar(200)"`
	ExampleSentence            string     `gorm:"column:example_sentence;type:varchar(500)"`
	VoiceRoleID                string     `gorm:"column:voice_role_id;type:varchar(36)"`
	WordAudioURL               *string    `gorm:"column:word_audio_url;type:varchar(500)"`
	WordAudioLocalFilename     *string    `gorm:"column:word_audio_local_filename;type:varchar(200)"`
	WordAudioGeneratedAt       *time.Time `gorm:"column:word_audio_generated_at"`
	SentenceAudioURL           *string    `gorm:"column:sentence_audio_url;type:varchar(500)"`
	SentenceAudioLocalFilename *string    `gorm:"column:sentence_audio_local_filename;type:varchar(200)"`
	SentenceAudioGeneratedAt   *time.Time `gorm:"column:sentence_audio_generated_at"`
	SortOrder                  int        `gorm:"column:sort_order"`
	Status                     string     `gorm:"column:status;type:varchar(20)"`
	Remark                     *string    `gorm:"column:remark;type:varchar(500)"`
	CreatedAt                  time.Time  `gorm:"column:created_at"`
	UpdatedAt                  time.Time  `gorm:"column:updated_at"`
}

func (ReadAloudVocabulary) TableName() string { return "sys_read_aloud_vocabularies" }

type ReadAloudSession struct {
	ID             string     `gorm:"column:id;type:varchar(36);primaryKey"`
	UserID         string     `gorm:"column:user_id;type:varchar(36)"`
	CategoryID     string     `gorm:"column:category_id;type:varchar(36)"`
	LanguageID     string     `gorm:"column:language_id;type:varchar(36)"`
	Status         string     `gorm:"column:status;type:varchar(20)"`
	TotalItems     int        `gorm:"column:total_items"`
	CompletedItems int        `gorm:"column:completed_items"`
	AverageScore   *int       `gorm:"column:average_score"`
	StartedAt      time.Time  `gorm:"column:started_at"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (ReadAloudSession) TableName() string { return "usr_read_aloud_sessions" }

type ReadAloudAttempt struct {
	ID            string    `gorm:"column:id;type:varchar(36);primaryKey"`
	SessionID     string    `gorm:"column:session_id;type:varchar(36)"`
	VocabularyID  string    `gorm:"column:vocabulary_id;type:varchar(36)"`
	Part          string    `gorm:"column:part;type:varchar(20)"`
	ReferenceText string    `gorm:"column:reference_text;type:varchar(500)"`
	Transcript    string    `gorm:"column:transcript;type:text"`
	Score         int       `gorm:"column:score"`
	MatchDetail   *string   `gorm:"column:match_detail;type:text"`
	DurationMs    *int      `gorm:"column:duration_ms"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (ReadAloudAttempt) TableName() string { return "usr_read_aloud_attempts" }

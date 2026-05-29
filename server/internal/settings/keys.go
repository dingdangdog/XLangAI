// Package settings 定义 sys_system_settings 的 key 约定与默认值（与 manager 种子、后台说明一致）。
package settings

// 认证：是否允许该方式登录 / 注册（与短信网关配置无关）。
const (
	AuthPasswordEnabled         = "auth.password.enabled"
	AuthPasswordRegisterEnabled = "auth.password.register_enabled"
	AuthSmsEnabled              = "auth.sms.enabled"
	AuthSmsRegisterEnabled      = "auth.sms.register_enabled"
)

// 媒体存储策略：client | server | cloud（与 sys_object_storage_configs 无关）。
const (
	MediaUserRecordingStorage = "media.user_recording.storage"
	MediaAssistantTtsStorage  = "media.assistant_tts.storage"
	MediaAvatarStorage        = "media.avatar.storage"
)

const (
	ValueTypeBool   = "bool"
	ValueTypeString = "string"
)

// Defaults 在库中无该行时使用的默认值。
var Defaults = map[string]string{
	AuthPasswordEnabled:         "true",
	AuthPasswordRegisterEnabled: "true",
	AuthSmsEnabled:              "true",
	AuthSmsRegisterEnabled:      "false",
	MediaUserRecordingStorage:   "server",
	MediaAssistantTtsStorage:    "server",
	MediaAvatarStorage:          "server",
}

// PublicKeys 可通过 GET /api/v1/public/settings 暴露给客户端的 key（不含任何密钥）。
var PublicKeys = []string{
	AuthPasswordEnabled,
	AuthPasswordRegisterEnabled,
	AuthSmsEnabled,
	AuthSmsRegisterEnabled,
	MediaUserRecordingStorage,
	MediaAssistantTtsStorage,
	MediaAvatarStorage,
}

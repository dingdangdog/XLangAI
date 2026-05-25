package translate

import (
	"encoding/json"
	"strings"
)

// ProviderConfig 从 sys_translate_service_configs.config JSON 解析的扩展字段。
type ProviderConfig struct {
	Region              string `json:"region"`
	APIVersion          string `json:"api_version"`
	AppID               string `json:"app_id"`
	ProjectID           string `json:"project_id"`
	SystemPromptTemplate string `json:"system_prompt_template"`
	MaxChars            int    `json:"max_chars"`
	UseFreeAPI          bool   `json:"use_free_api"`
}

func ParseProviderConfig(raw string) ProviderConfig {
	var c ProviderConfig
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return c
	}
	_ = json.Unmarshal([]byte(raw), &c)
	return c
}

// ServiceInput 调用翻译路由时的运行时参数。
type ServiceInput struct {
	Protocol   string
	BaseURL    string
	APIKey     string
	APISecret  string
	ModelCode  string
	LlmConfigID string
	Config     ProviderConfig
}

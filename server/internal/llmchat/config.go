package llmchat

import (
	"encoding/json"
	"strings"
)

// ProviderConfig 从 sys_llm_service_configs.config JSON 解析。
type ProviderConfig struct {
	AnthropicVersion string  `json:"anthropic_version"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
}

// ServiceInput 对话 Chat 路由参数。
type ServiceInput struct {
	Protocol     string
	BaseURL      string
	APIKey       string
	ModelCode    string
	Config       ProviderConfig
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

// IsOpenAICompatible 使用 OpenAI Chat Completions 形态的协议（含多数兼容网关）。
func IsOpenAICompatible(protocol string) bool {
	switch strings.TrimSpace(strings.ToLower(protocol)) {
	case "", "openai", "azure_openai", "ollama", "deepseek", "openrouter", "groq",
		"together", "zhipu", "moonshot", "siliconflow", "nvidia_nim", "mistral":
		return true
	default:
		return false
	}
}

package ai

import (
	"encoding/json"
	"strings"
)

// IsVolcengineArkBaseURL 火山方舟 OpenAI 兼容网关（路径为 /api/v3，非 /v1）。
func IsVolcengineArkBaseURL(baseURL string) bool {
	u := strings.ToLower(strings.TrimSpace(baseURL))
	return strings.Contains(u, "volces.com") || strings.Contains(u, "volcengine.com")
}

// OpenAIChatCompletionsURL 按网关形态拼接 Chat Completions 完整 URL。
func OpenAIChatCompletionsURL(baseURL string) string {
	root := strings.TrimRight(NormalizeOpenAIBaseURL(baseURL), "/")
	if IsVolcengineArkBaseURL(root) {
		return root + "/chat/completions"
	}
	return root + "/v1/chat/completions"
}

// OpenAIModelsURL 按网关形态拼接模型列表 URL。
func OpenAIModelsURL(baseURL string) string {
	root := strings.TrimRight(NormalizeOpenAIBaseURL(baseURL), "/")
	if IsVolcengineArkBaseURL(root) {
		return root + "/models"
	}
	return root + "/v1/models"
}

// ChatOptions OpenAI 兼容 Chat 的可选参数。
type ChatOptions struct {
	MaxTokens           int
	MaxCompletionTokens int
	Temperature         float64
	TopP                float64
	// ThinkingJSON 方舟 thinking 字段原始 JSON，如 {"type":"disabled"}。
	ThinkingJSON json.RawMessage
}

// BuildChatCompletionBody 构造 chat.completions 请求体（含方舟专有字段）。
func BuildChatCompletionBody(model string, msgs []chatMessage, opts *ChatOptions, baseURL string) ([]byte, error) {
	body := map[string]any{
		"model":    model,
		"messages": msgs,
	}
	if opts != nil {
		if opts.Temperature > 0 {
			body["temperature"] = opts.Temperature
		}
		if opts.TopP > 0 {
			body["top_p"] = opts.TopP
		}
		tokenParams := buildChatCompletionTokenParams(baseURL, opts)
		for k, v := range tokenParams {
			body[k] = v
		}
		extras := buildOpenAICompatibleChatBodyExtras(baseURL, opts)
		for k, v := range extras {
			body[k] = v
		}
	} else if IsVolcengineArkBaseURL(baseURL) {
		for k, v := range buildOpenAICompatibleChatBodyExtras(baseURL, nil) {
			body[k] = v
		}
	}
	return json.Marshal(body)
}

func buildChatCompletionTokenParams(baseURL string, opts *ChatOptions) map[string]int {
	raw := 0
	if opts != nil {
		if opts.MaxCompletionTokens > 0 {
			raw = opts.MaxCompletionTokens
		} else if opts.MaxTokens > 0 {
			raw = opts.MaxTokens
		}
	}
	if raw <= 0 {
		raw = 4096
	}
	if raw < 16 {
		raw = 16
	}
	if raw > 32000 {
		raw = 32000
	}
	if IsVolcengineArkBaseURL(baseURL) {
		return map[string]int{"max_completion_tokens": raw}
	}
	return map[string]int{"max_tokens": raw}
}

func buildOpenAICompatibleChatBodyExtras(baseURL string, opts *ChatOptions) map[string]any {
	if opts != nil && len(opts.ThinkingJSON) > 0 {
		var thinking any
		if err := json.Unmarshal(opts.ThinkingJSON, &thinking); err == nil {
			return map[string]any{"thinking": thinking}
		}
	}
	if !IsVolcengineArkBaseURL(baseURL) {
		return nil
	}
	return map[string]any{"thinking": map[string]string{"type": "disabled"}}
}

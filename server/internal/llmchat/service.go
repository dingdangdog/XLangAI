package llmchat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"xlangai/server/internal/ai"
)

var (
	ErrProviderNotReady    = errors.New("llm provider not configured")
	ErrUnsupportedProtocol = errors.New("unsupported llm protocol")
)

func truncateErr(b []byte, max int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

// Chat 按 protocol 路由到对应 LLM 服务商。
func Chat(
	ctx context.Context,
	in ServiceInput,
	systemPrompt string,
	messages []Message,
) (string, *ai.ChatUsage, error) {
	protocol := strings.TrimSpace(strings.ToLower(in.Protocol))
	if protocol == "" {
		protocol = "openai"
	}
	if IsOpenAICompatible(protocol) {
		return chatOpenAI(ctx, in, systemPrompt, messages)
	}
	switch protocol {
	case "claude", "anthropic":
		return chatClaude(ctx, in, systemPrompt, messages)
	case "gemini", "google_gemini":
		return chatGemini(ctx, in, systemPrompt, messages)
	default:
		return "", nil, fmt.Errorf("%w: %s", ErrUnsupportedProtocol, protocol)
	}
}

// ServiceInputFromRepo 由仓库配置构造 ServiceInput。
func ServiceInputFromRepo(protocol, baseURL, apiKey, modelCode, configRaw string) ServiceInput {
	return ServiceInput{
		Protocol:  protocol,
		BaseURL:   baseURL,
		APIKey:    apiKey,
		ModelCode: modelCode,
		Config:    ParseProviderConfig(configRaw),
	}
}

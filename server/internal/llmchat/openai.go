package llmchat

import (
	"context"
	"strings"

	"xlangai/server/internal/ai"
)

func chatOpenAI(ctx context.Context, in ServiceInput, systemPrompt string, messages []Message) (string, *ai.ChatUsage, error) {
	apiKey := strings.TrimSpace(in.APIKey)
	if apiKey == "" {
		return "", nil, ErrProviderNotReady
	}
	baseURL := in.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	client := ai.NewOpenAIClient(apiKey, in.ModelCode, baseURL)
	msgs := make([]struct{ Role, Content string }, len(messages))
	for i, m := range messages {
		msgs[i] = struct{ Role, Content string }{Role: m.Role, Content: m.Content}
	}
	return client.Chat(ctx, systemPrompt, msgs)
}

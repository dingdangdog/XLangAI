package translate

import (
	"context"
	"fmt"
	"strings"

	"xlangai/server/internal/ai"
)

func translateOpenAI(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	apiKey := strings.TrimSpace(in.APIKey)
	if apiKey == "" {
		return "", ErrProviderNotReady
	}
	baseURL := strings.TrimSpace(in.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	model := strings.TrimSpace(in.ModelCode)
	if model == "" {
		model = "gpt-4o-mini"
	}
	client := ai.NewOpenAIClient(apiKey, model, baseURL)

	tpl := strings.TrimSpace(in.Config.SystemPromptTemplate)
	if tpl == "" {
		tpl = "You are a professional translator. Translate the following message into the language that best matches the BCP-47 / IETF locale tag: %s. " +
			"Preserve meaning, tone, and line breaks where reasonable. Output ONLY the translated text with no quotation marks, labels, or explanations."
	}
	system := fmt.Sprintf(tpl, targetLocale)
	msgs := []struct{ Role, Content string }{{Role: "user", Content: text}}
	out, _, err := client.Chat(ctx, system, msgs)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

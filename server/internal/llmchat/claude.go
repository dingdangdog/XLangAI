package llmchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"wlltalk/server/internal/ai"
)

func chatClaude(ctx context.Context, in ServiceInput, systemPrompt string, messages []Message) (string, *ai.ChatUsage, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return "", nil, ErrProviderNotReady
	}
	base := strings.TrimSpace(in.BaseURL)
	if base == "" {
		base = "https://api.anthropic.com"
	}
	base = strings.TrimRight(base, "/")
	ver := strings.TrimSpace(in.Config.AnthropicVersion)
	if ver == "" {
		ver = "2023-06-01"
	}
	maxTok := in.Config.MaxTokens
	if maxTok <= 0 {
		maxTok = 4096
	}
	model := strings.TrimSpace(in.ModelCode)
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	apiMsgs := make([]msg, 0, len(messages))
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role == "assistant" {
			role = "assistant"
		} else if role != "user" {
			continue
		}
		apiMsgs = append(apiMsgs, msg{Role: role, Content: m.Content})
	}
	bodyMap := map[string]any{
		"model":      model,
		"max_tokens": maxTok,
		"messages":   apiMsgs,
	}
	if sp := strings.TrimSpace(systemPrompt); sp != "" {
		bodyMap["system"] = sp
	}
	if in.Config.Temperature > 0 {
		bodyMap["temperature"] = in.Config.Temperature
	}
	if in.Config.TopP > 0 {
		bodyMap["top_p"] = in.Config.TopP
	}
	body, _ := json.Marshal(bodyMap)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", ver)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("claude api error %d: %s", resp.StatusCode, truncateErr(b, 512))
	}
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", nil, err
	}
	var out strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" && block.Text != "" {
			out.WriteString(block.Text)
		}
	}
	text := strings.TrimSpace(out.String())
	if text == "" {
		return "", nil, fmt.Errorf("claude: empty response")
	}
	usage := &ai.ChatUsage{
		PromptTokens:     parsed.Usage.InputTokens,
		CompletionTokens: parsed.Usage.OutputTokens,
		TotalTokens:      parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
	}
	return text, usage, nil
}

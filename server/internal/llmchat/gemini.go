package llmchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"xlangai/server/internal/ai"
)

func chatGemini(ctx context.Context, in ServiceInput, systemPrompt string, messages []Message) (string, *ai.ChatUsage, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return "", nil, ErrProviderNotReady
	}
	base := strings.TrimSpace(in.BaseURL)
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	base = strings.TrimRight(base, "/")
	model := strings.TrimSpace(in.ModelCode)
	if model == "" {
		model = "gemini-1.5-flash"
	}
	model = strings.TrimPrefix(model, "models/")
	model = strings.TrimPrefix(model, "models\\")

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}
	contents := make([]content, 0, len(messages))
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		switch role {
		case "assistant":
			role = "model"
		case "user":
			role = "user"
		default:
			continue
		}
		contents = append(contents, content{
			Role:  role,
			Parts: []part{{Text: m.Content}},
		})
	}
	reqBody := map[string]any{"contents": contents}
	if sp := strings.TrimSpace(systemPrompt); sp != "" {
		reqBody["systemInstruction"] = map[string]any{
			"parts": []part{{Text: sp}},
		}
	}
	genCfg := map[string]any{}
	if in.Config.Temperature > 0 {
		genCfg["temperature"] = in.Config.Temperature
	}
	if in.Config.TopP > 0 {
		genCfg["topP"] = in.Config.TopP
	}
	if in.Config.MaxTokens > 0 {
		genCfg["maxOutputTokens"] = in.Config.MaxTokens
	}
	if len(genCfg) > 0 {
		reqBody["generationConfig"] = genCfg
	}
	body, _ := json.Marshal(reqBody)
	reqURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", base, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("gemini api error %d: %s", resp.StatusCode, truncateErr(b, 512))
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", nil, err
	}
	var out strings.Builder
	if len(parsed.Candidates) > 0 {
		for _, p := range parsed.Candidates[0].Content.Parts {
			out.WriteString(p.Text)
		}
	}
	text := strings.TrimSpace(out.String())
	if text == "" {
		return "", nil, fmt.Errorf("gemini: empty response")
	}
	usage := &ai.ChatUsage{
		PromptTokens:     parsed.UsageMetadata.PromptTokenCount,
		CompletionTokens: parsed.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      parsed.UsageMetadata.TotalTokenCount,
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return text, usage, nil
}

package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GeminiTTSSpeech 使用 Gemini generateContent 的 AUDIO 模态合成语音。
// voiceName 来自语音角色 voice_code（如 Kore、Charon）；model 来自 model_code。
func GeminiTTSSpeech(ctx context.Context, apiKey, baseURL, model, voiceName, text string) ([]byte, string, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, "", fmt.Errorf("gemini tts: missing api key")
	}
	if strings.TrimSpace(text) == "" {
		return nil, "", fmt.Errorf("gemini tts: empty text")
	}
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	base = strings.TrimRight(base, "/")
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-2.5-flash-preview-tts"
	}
	model = strings.TrimPrefix(model, "models/")
	voice := strings.TrimSpace(voiceName)
	if voice == "" {
		voice = "Kore"
	}
	reqBody := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": text}}},
		},
		"generationConfig": map[string]any{
			"responseModalities": []string{"AUDIO"},
			"speechConfig": map[string]any{
				"voiceConfig": map[string]any{
					"prebuiltVoiceConfig": map[string]string{
						"voiceName": voice,
					},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	reqURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", base, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("gemini tts error %d: %s", resp.StatusCode, truncateTTSErr(b))
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, "", err
	}
	for _, c := range parsed.Candidates {
		for _, p := range c.Content.Parts {
			if p.InlineData == nil || p.InlineData.Data == "" {
				continue
			}
			raw, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
			if err != nil {
				return nil, "", err
			}
			mime := strings.TrimSpace(p.InlineData.MimeType)
			if mime == "" {
				mime = "audio/mpeg"
			}
			return raw, mime, nil
		}
	}
	return nil, "", fmt.Errorf("gemini tts: no audio in response")
}

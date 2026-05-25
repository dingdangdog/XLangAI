package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ElevenLabsTTS POST /v1/text-to-speech/{voice_id}
func ElevenLabsTTS(ctx context.Context, apiKey, baseURL, voiceID, modelID, text string) ([]byte, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("elevenlabs tts: missing api key")
	}
	voice := strings.TrimSpace(voiceID)
	if voice == "" {
		return nil, fmt.Errorf("elevenlabs tts: empty voice_id")
	}
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "https://api.elevenlabs.io"
	}
	base = strings.TrimRight(base, "/")
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = "eleven_multilingual_v2"
	}
	body, _ := json.Marshal(map[string]any{
		"text":     text,
		"model_id": model,
	})
	reqURL := base + "/v1/text-to-speech/" + voice
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", key)
	req.Header.Set("Accept", "audio/mpeg")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("elevenlabs tts: %s: %s", resp.Status, truncateTTSErr(data))
	}
	return data, nil
}

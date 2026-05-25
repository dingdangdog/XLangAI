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

// DeepgramTTS Speak API。
func DeepgramTTS(ctx context.Context, apiKey, model, text string) ([]byte, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("deepgram tts: missing api key")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "aura-asteria-en"
	}
	body, _ := json.Marshal(map[string]string{"text": text})
	reqURL := "https://api.deepgram.com/v1/speak?model=" + model
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+key)
	req.Header.Set("Accept", "audio/mpeg")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("deepgram tts: %s: %s", resp.Status, truncateTTSErr(data))
	}
	return data, nil
}

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

// MinimaxTTS MiniMax 语音合成 t2a_v2。
func MinimaxTTS(ctx context.Context, apiKey, baseURL, model, voiceID, text string) ([]byte, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("minimax tts: missing api key")
	}
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "https://api.minimaxi.com"
	}
	base = strings.TrimRight(base, "/")
	if model == "" {
		model = "speech-02-turbo"
	}
	voice := strings.TrimSpace(voiceID)
	if voice == "" {
		voice = "female-shaonv"
	}
	body, _ := json.Marshal(map[string]any{
		"model": model,
		"text":  text,
		"voice_setting": map[string]any{
			"voice_id": voice,
			"speed":    1.0,
			"vol":      1.0,
			"pitch":    0,
		},
		"audio_setting": map[string]any{
			"format": "mp3",
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/t2a_v2", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("minimax tts: %s: %s", resp.Status, truncateTTSErr(b))
	}
	var parsed struct {
		Data struct {
			Audio string `json:"audio"`
		} `json:"data"`
		BaseResp struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if parsed.BaseResp.StatusCode != 0 && parsed.BaseResp.StatusCode != 200 {
		return nil, fmt.Errorf("minimax tts: %s", parsed.BaseResp.StatusMsg)
	}
	if parsed.Data.Audio == "" {
		return nil, fmt.Errorf("minimax tts: empty audio")
	}
	return base64.StdEncoding.DecodeString(parsed.Data.Audio)
}

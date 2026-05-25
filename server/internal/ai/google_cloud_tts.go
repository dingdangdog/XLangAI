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

// GoogleCloudTTSSpeech Cloud Text-to-Speech REST（与 Gemini TTS 不同）。
// voice_code 填完整 voice 名（如 en-US-Neural2-A）；language_code 可在 config 中指定。
func GoogleCloudTTSSpeech(ctx context.Context, apiKey, voiceName, languageCode, encoding, text string) ([]byte, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("google cloud tts: missing api key")
	}
	voice := strings.TrimSpace(voiceName)
	if voice == "" {
		voice = "en-US-Neural2-A"
	}
	lang := strings.TrimSpace(languageCode)
	if lang == "" {
		parts := strings.Split(voice, "-")
		if len(parts) >= 2 {
			lang = parts[0] + "-" + parts[1]
		} else {
			lang = "en-US"
		}
	}
	if encoding == "" {
		encoding = "MP3"
	}
	body, _ := json.Marshal(map[string]any{
		"input": map[string]string{"text": text},
		"voice": map[string]string{"languageCode": lang, "name": voice},
		"audioConfig": map[string]string{"audioEncoding": encoding},
	})
	reqURL := "https://texttospeech.googleapis.com/v1/text:synthesize?key=" + key
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google cloud tts: %s: %s", resp.Status, truncateTTSErr(b))
	}
	var parsed struct {
		AudioContent string `json:"audioContent"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if parsed.AudioContent == "" {
		return nil, fmt.Errorf("google cloud tts: empty audio")
	}
	return base64.StdEncoding.DecodeString(parsed.AudioContent)
}

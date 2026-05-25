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

// OpenAITTSSpeech 调用 OpenAI 兼容的 /v1/audio/speech，返回 mp3 字节。
// baseURL 应为 NormalizeOpenAIBaseURL 处理后的根地址。
func OpenAITTSSpeech(ctx context.Context, baseURL, apiKey, model, voice, text string) ([]byte, error) {
	url := strings.TrimRight(baseURL, "/") + "/v1/audio/speech"

	body, _ := json.Marshal(map[string]interface{}{
		"model": model,
		"input": text,
		"voice": voice,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts api error %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}

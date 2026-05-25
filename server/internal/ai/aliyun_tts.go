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

// AliyunNLSTTS 阿里云智能语音交互 REST 流式 TTS（需有效 X-NLS-Token + appkey）。
// voice 使用语音角色 voice_code（如 xiaoyun）；region 为网关区域（cn-shanghai 等）。
func AliyunNLSTTS(ctx context.Context, token, appKey, region, voice, format string, sampleRate int, text string) ([]byte, string, error) {
	token = strings.TrimSpace(token)
	appKey = strings.TrimSpace(appKey)
	if token == "" || appKey == "" {
		return nil, "", fmt.Errorf("aliyun tts: missing token or app_key")
	}
	if strings.TrimSpace(voice) == "" {
		return nil, "", fmt.Errorf("aliyun tts: empty voice")
	}
	region = strings.TrimSpace(region)
	if region == "" {
		region = "cn-shanghai"
	}
	if format == "" {
		format = "mp3"
	}
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	endpoint := fmt.Sprintf("https://nls-gateway-%s.aliyuncs.com/stream/v1/tts", strings.ToLower(region))
	bodyMap := map[string]any{
		"appkey":      appKey,
		"text":        text,
		"format":      format,
		"sample_rate": sampleRate,
		"voice":       voice,
	}
	body, _ := json.Marshal(bodyMap)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-NLS-Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("aliyun tts error %d: %s", resp.StatusCode, truncateTTSErr(data))
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "audio/mpeg"
	}
	return data, mime, nil
}

func truncateTTSErr(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 400 {
		return s[:400] + "…"
	}
	return s
}

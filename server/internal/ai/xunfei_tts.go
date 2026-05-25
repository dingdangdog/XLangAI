package ai

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// XunfeiTTS 讯飞开放平台在线语音合成 WebAPI v2。
// config: app_id, api_secret；apiKey 为 api_key；voice_code 为发音人（如 xiaoyan）。
func XunfeiTTS(ctx context.Context, appID, apiKey, apiSecret, voice, text string) ([]byte, error) {
	appID = strings.TrimSpace(appID)
	apiKey = strings.TrimSpace(apiKey)
	apiSecret = strings.TrimSpace(apiSecret)
	if appID == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("xunfei tts: missing app_id, api_key or api_secret")
	}
	if voice == "" {
		voice = "xiaoyan"
	}
	host := "tts-api.xfyun.cn"
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	signString := fmt.Sprintf("host: %s\ndate: %s\nPOST /v2/tts HTTP/1.1", host, date)
	mac := hmac.New(sha256.New, []byte(apiSecret))
	_, _ = mac.Write([]byte(signString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`, apiKey, signature)))
	body, _ := json.Marshal(map[string]any{
		"common": map[string]string{"app_id": appID},
		"business": map[string]any{
			"aue": "lame", "sfl": 1, "auf": "audio/L16;rate=16000",
			"vcn": voice, "speed": 50, "volume": 50, "pitch": 50, "tte": "UTF8",
		},
		"data": map[string]any{
			"status": 2,
			"text":   base64.StdEncoding.EncodeToString([]byte(text)),
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+host+"/v2/tts", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("xunfei tts: %s: %s", resp.Status, truncateTTSErr(b))
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Audio string `json:"audio"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("xunfei tts: %d %s", parsed.Code, parsed.Message)
	}
	if parsed.Data.Audio == "" {
		return nil, fmt.Errorf("xunfei tts: empty audio")
	}
	return base64.StdEncoding.DecodeString(parsed.Data.Audio)
}

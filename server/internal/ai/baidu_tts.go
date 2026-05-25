package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// BaiduTTSSpeech 百度语音合成（需有效 access_token 放在 apiKey）。
// voice_code 为 per 发音人编号（如 0、1、3、4）；config 可含 spd/pit/vol。
func BaiduTTSSpeech(ctx context.Context, accessToken, cuid, per string, spd, pit, vol int, text string) ([]byte, error) {
	token := strings.TrimSpace(accessToken)
	if token == "" {
		return nil, fmt.Errorf("baidu tts: missing access_token")
	}
	if per == "" {
		per = "0"
	}
	if cuid == "" {
		cuid = "xlangai"
	}
	q := url.Values{}
	q.Set("tex", text)
	q.Set("tok", token)
	q.Set("cuid", cuid)
	q.Set("ctp", "1")
	q.Set("lan", "zh")
	q.Set("per", per)
	if spd > 0 {
		q.Set("spd", strconv.Itoa(spd))
	}
	if pit > 0 {
		q.Set("pit", strconv.Itoa(pit))
	}
	if vol > 0 {
		q.Set("vol", strconv.Itoa(vol))
	}
	reqURL := "https://tsn.baidu.com/text2audio?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || strings.Contains(ct, "json") {
		return nil, fmt.Errorf("baidu tts: %s: %s", resp.Status, truncateTTSErr(data))
	}
	return data, nil
}

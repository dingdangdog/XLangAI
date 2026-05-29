package translate

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

func translateXunfei(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	appID := strings.TrimSpace(in.Config.AppID)
	apiKey := strings.TrimSpace(in.APIKey)
	apiSecret := strings.TrimSpace(in.APISecret)
	if appID == "" || apiKey == "" || apiSecret == "" {
		return "", ErrProviderNotReady
	}
	from, to := xunfeiLangPair(targetLocale)
	host := "ntrans.xfyun.cn"
	endpoint := "https://" + host + "/v2/ots"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
		if u := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://"); u != "" {
			if i := strings.IndexByte(u, '/'); i >= 0 {
				host = u[:i]
			}
		}
	}
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	signString := fmt.Sprintf("host: %s\ndate: %s\nPOST /v2/ots HTTP/1.1", host, date)
	mac := hmac.New(sha256.New, []byte(apiSecret))
	_, _ = mac.Write([]byte(signString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
		`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`,
		apiKey, signature,
	)))
	body, _ := json.Marshal(map[string]any{
		"common": map[string]string{"app_id": appID},
		"business": map[string]string{"from": from, "to": to},
		"data": map[string]any{
			"text": base64.StdEncoding.EncodeToString([]byte(text)),
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("xunfei translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Result struct {
				TransResult struct {
					Dst string `json:"dst"`
				} `json:"trans_result"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("xunfei translate: %d %s", parsed.Code, parsed.Message)
	}
	out := strings.TrimSpace(parsed.Data.Result.TransResult.Dst)
	if out == "" {
		return "", fmt.Errorf("xunfei translate: empty result")
	}
	return out, nil
}

func xunfeiLangPair(targetLocale string) (from, to string) {
	to = TargetForProvider("xunfei_translate", targetLocale)
	return "auto", to
}

package translate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func translateYoudao(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	appKey := strings.TrimSpace(in.APIKey)
	appSecret := strings.TrimSpace(in.APISecret)
	if appKey == "" || appSecret == "" {
		return "", ErrProviderNotReady
	}
	to := TargetForProvider("youdao_translate", targetLocale)
	from := "auto"
	endpoint := "https://openapi.youdao.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	salt := strconv.FormatInt(time.Now().UnixNano(), 10)
	curtime := strconv.FormatInt(time.Now().Unix(), 10)
	sign := youdaoSignV3(appKey, text, salt, curtime, appSecret)
	form := url.Values{}
	form.Set("q", text)
	form.Set("from", from)
	form.Set("to", to)
	form.Set("appKey", appKey)
	form.Set("salt", salt)
	form.Set("sign", sign)
	form.Set("signType", "v3")
	form.Set("curtime", curtime)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/api", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var parsed struct {
		ErrorCode   string `json:"errorCode"`
		Translation []string `json:"translation"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.ErrorCode != "" && parsed.ErrorCode != "0" {
		return "", fmt.Errorf("youdao translate: %s", parsed.ErrorCode)
	}
	if len(parsed.Translation) == 0 {
		return "", fmt.Errorf("youdao translate: empty result")
	}
	return strings.TrimSpace(parsed.Translation[0]), nil
}

func youdaoSignV3(appKey, q, salt, curtime, appSecret string) string {
	input := youdaoInput(q)
	raw := appKey + input + salt + curtime + appSecret
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func youdaoInput(q string) string {
	n := len([]rune(q))
	if n <= 20 {
		return q
	}
	runes := []rune(q)
	return string(runes[:10]) + strconv.Itoa(n) + string(runes[n-10:])
}

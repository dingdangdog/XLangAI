package translate

import (
	"context"
	"crypto/md5"
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

func translateBaidu(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	appID := strings.TrimSpace(in.APIKey)
	secret := strings.TrimSpace(in.APISecret)
	if appID == "" {
		appID = strings.TrimSpace(in.Config.AppID)
	}
	if secret == "" {
		secret = strings.TrimSpace(in.APIKey)
	}
	if appID == "" || secret == "" {
		return "", ErrProviderNotReady
	}
	to := TargetForProvider("baidu_translate", targetLocale)
	from := "auto"
	endpoint := "https://fanyi-api.baidu.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	salt := strconv.FormatInt(time.Now().UnixNano(), 10)
	sign := md5Hex(appID + text + salt + secret)
	q := url.Values{}
	q.Set("q", text)
	q.Set("from", from)
	q.Set("to", to)
	q.Set("appid", appID)
	q.Set("salt", salt)
	q.Set("sign", sign)
	reqURL := endpoint + "/api/trans/vip/translate?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var parsed struct {
		ErrorCode string `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
		TransResult []struct {
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.ErrorCode != "" && parsed.ErrorCode != "0" {
		return "", fmt.Errorf("baidu translate: %s %s", parsed.ErrorCode, parsed.ErrorMsg)
	}
	if len(parsed.TransResult) == 0 {
		return "", fmt.Errorf("baidu translate: empty result")
	}
	return strings.TrimSpace(parsed.TransResult[0].Dst), nil
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

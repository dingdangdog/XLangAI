package translate

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func translateAliyun(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	accessKeyID := strings.TrimSpace(in.APIKey)
	accessKeySecret := strings.TrimSpace(in.APISecret)
	if accessKeyID == "" {
		accessKeyID = strings.TrimSpace(in.Config.AppID)
	}
	if accessKeySecret == "" {
		accessKeySecret = strings.TrimSpace(in.APIKey)
	}
	if accessKeyID == "" || accessKeySecret == "" {
		return "", ErrProviderNotReady
	}
	region := strings.TrimSpace(in.Config.Region)
	if region == "" {
		region = "cn-hangzhou"
	}
	endpoint := "mt.aliyuncs.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" {
		bu = strings.TrimPrefix(strings.TrimPrefix(bu, "https://"), "http://")
		bu = strings.TrimRight(bu, "/")
		if bu != "" {
			endpoint = bu
		}
	}
	to := TargetForProvider("aliyun_translate", targetLocale)
	params := map[string]string{
		"Action":           "TranslateGeneral",
		"Format":           "JSON",
		"Version":          "2018-10-12",
		"AccessKeyId":      accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"FormatType":       "text",
		"SourceLanguage":   "auto",
		"TargetLanguage":   to,
		"SourceText":       text,
		"RegionId":         region,
	}
	sign := aliyunSign(params, accessKeySecret, "POST")
	params["Signature"] = sign
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+endpoint+"/", strings.NewReader(form.Encode()))
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("aliyun translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		Data    struct {
			Translated string `json:"Translated"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != "" && parsed.Code != "200" {
		return "", fmt.Errorf("aliyun translate: %s %s", parsed.Code, parsed.Message)
	}
	out := strings.TrimSpace(parsed.Data.Translated)
	if out == "" {
		return "", fmt.Errorf("aliyun translate: empty result")
	}
	return out, nil
}

func aliyunSign(params map[string]string, secret, method string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonical bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			canonical.WriteByte('&')
		}
		canonical.WriteString(percentEncode(k))
		canonical.WriteByte('=')
		canonical.WriteString(percentEncode(params[k]))
	}
	stringToSign := method + "&%2F&" + percentEncode(canonical.String())
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func percentEncode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

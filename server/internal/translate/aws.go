package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func translateAWS(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	accessKeyID := strings.TrimSpace(in.APIKey)
	secretAccessKey := strings.TrimSpace(in.APISecret)
	if accessKeyID == "" || secretAccessKey == "" {
		return "", ErrProviderNotReady
	}
	region := strings.TrimSpace(in.Config.Region)
	if region == "" {
		region = "us-east-1"
	}
	host := "translate." + region + ".amazonaws.com"
	endpoint := "https://" + host + "/"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/") + "/"
		if u := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://"); u != "" {
			if i := strings.IndexByte(u, '/'); i >= 0 {
				host = u[:i]
			} else {
				host = strings.TrimRight(u, "/")
			}
		}
	}
	to := TargetForProvider("aws_translate", targetLocale)
	body, _ := json.Marshal(map[string]string{
		"Text":               text,
		"SourceLanguageCode": "auto",
		"TargetLanguageCode": to,
	})
	payload := string(body)
	extra := map[string]string{"x-amz-target": "AWSShineFrontendService_20170701.TranslateText"}
	auth, amzDate, err := aws4Authorization(accessKeyID, secretAccessKey, region, "translate", host, "POST", "/", payload, extra)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("Host", host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Target", extra["x-amz-target"])
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("aws translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		TranslatedText string `json:"TranslatedText"`
		Message        string `json:"message"`
		Reason         string `json:"__type"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	out := strings.TrimSpace(parsed.TranslatedText)
	if out == "" {
		msg := parsed.Message
		if msg == "" {
			msg = parsed.Reason
		}
		return "", fmt.Errorf("aws translate: empty result: %s", msg)
	}
	return out, nil
}

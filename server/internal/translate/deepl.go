package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func translateDeepL(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return "", ErrProviderNotReady
	}
	endpoint := "https://api.deepl.com"
	if in.Config.UseFreeAPI {
		endpoint = "https://api-free.deepl.com"
	}
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	to := TargetForProvider("deepl", targetLocale)
	form := url.Values{}
	form.Set("text", text)
	form.Set("target_lang", to)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/v2/translate", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("deepl: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Translations) == 0 {
		return "", fmt.Errorf("deepl: empty result")
	}
	return strings.TrimSpace(parsed.Translations[0].Text), nil
}

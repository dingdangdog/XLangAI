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

func translateGoogle(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return "", ErrProviderNotReady
	}
	to := TargetForProvider("google_translate", targetLocale)
	endpoint := "https://translation.googleapis.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	q := url.Values{}
	q.Set("key", key)
	q.Set("q", text)
	q.Set("target", to)
	q.Set("format", "text")
	reqURL := endpoint + "/language/translate/v2?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("google translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Data struct {
			Translations []struct {
				TranslatedText string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Data.Translations) == 0 {
		return "", fmt.Errorf("google translate: empty result")
	}
	return strings.TrimSpace(parsed.Data.Translations[0].TranslatedText), nil
}

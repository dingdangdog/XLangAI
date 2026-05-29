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

func translateLibre(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	endpoint := strings.TrimSpace(in.BaseURL)
	if endpoint == "" {
		endpoint = "http://localhost:5000"
	}
	if !strings.HasPrefix(strings.ToLower(endpoint), "http") {
		endpoint = "http://" + endpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")
	to := TargetForProvider("libretranslate", targetLocale)
	payload := map[string]string{
		"q":      text,
		"source": "auto",
		"target": to,
		"format": "text",
	}
	if key := strings.TrimSpace(in.APIKey); key != "" {
		payload["api_key"] = key
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/translate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("libretranslate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		TranslatedText string `json:"translatedText"`
		Error          string `json:"error"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("libretranslate: %s", parsed.Error)
	}
	out := strings.TrimSpace(parsed.TranslatedText)
	if out == "" {
		return "", fmt.Errorf("libretranslate: empty result")
	}
	return out, nil
}

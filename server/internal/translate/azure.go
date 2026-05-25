package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func translateAzure(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		return "", ErrProviderNotReady
	}
	region := strings.TrimSpace(in.Config.Region)
	if region == "" {
		region = strings.TrimSpace(in.BaseURL)
	}
	to := TargetForProvider("azure_translator", targetLocale)
	ver := strings.TrimSpace(in.Config.APIVersion)
	if ver == "" {
		ver = "3.0"
	}
	endpoint := "https://api.cognitive.microsofttranslator.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	q := url.Values{}
	q.Set("api-version", ver)
	q.Set("to", to)
	reqURL := endpoint + "/translate?" + q.Encode()

	body, _ := json.Marshal([]map[string]string{{"Text": text}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", key)
	if region != "" && !strings.HasPrefix(strings.ToLower(region), "http") {
		req.Header.Set("Ocp-Apim-Subscription-Region", region)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("azure translator: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed []struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if len(parsed) == 0 || len(parsed[0].Translations) == 0 {
		return "", fmt.Errorf("azure translator: empty result")
	}
	return strings.TrimSpace(parsed[0].Translations[0].Text), nil
}

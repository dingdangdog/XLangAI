package translate

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func translateIBM(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	apiKey := strings.TrimSpace(in.APIKey)
	if apiKey == "" {
		return "", ErrProviderNotReady
	}
	region := strings.TrimSpace(in.Config.Region)
	if region == "" {
		region = "us-south"
	}
	endpoint := "https://api." + region + ".language-translator.watson.cloud.ibm.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	modelID := strings.TrimSpace(in.ModelCode)
	if modelID == "" || modelID == "-" {
		modelID = ibmModelID(targetLocale)
	}
	version := strings.TrimSpace(in.Config.APIVersion)
	if version == "" {
		version = "2018-05-01"
	}
	body, _ := json.Marshal(map[string]any{
		"text":     []string{text},
		"model_id": modelID,
	})
	reqURL := endpoint + "/v3/translate?version=" + version
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(apiKey+":")))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ibm translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Translations []struct {
			Translation string `json:"translation"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Translations) == 0 {
		return "", fmt.Errorf("ibm translate: empty result")
	}
	return strings.TrimSpace(parsed.Translations[0].Translation), nil
}

func ibmModelID(targetLocale string) string {
	to := TargetForProvider("ibm_watson_translate", targetLocale)
	if to == "zh" || strings.HasPrefix(strings.ToLower(to), "zh") {
		return "en-zh"
	}
	return "en-" + to
}

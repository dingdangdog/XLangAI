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

func translatePapago(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	clientID := strings.TrimSpace(in.APIKey)
	clientSecret := strings.TrimSpace(in.APISecret)
	if clientID == "" || clientSecret == "" {
		return "", ErrProviderNotReady
	}
	to := TargetForProvider("papago_translate", targetLocale)
	endpoint := "https://papago.apigw.ntruss.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	form := url.Values{}
	form.Set("source", "auto")
	form.Set("target", to)
	form.Set("text", text)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/nmt/v1/translation", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-NCP-APIGW-API-KEY-ID", clientID)
	req.Header.Set("X-NCP-APIGW-API-KEY", clientSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("papago translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Message struct {
			Result struct {
				Translated string `json:"translated"`
			} `json:"result"`
		} `json:"message"`
		Error struct {
			ErrorCode string `json:"errorCode"`
			Message   string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.Error.ErrorCode != "" {
		return "", fmt.Errorf("papago translate: %s %s", parsed.Error.ErrorCode, parsed.Error.Message)
	}
	out := strings.TrimSpace(parsed.Message.Result.Translated)
	if out == "" {
		return "", fmt.Errorf("papago translate: empty result")
	}
	return out, nil
}

package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// IBMWatsonTTS IBM Cloud Text to Speech synthesize。
func IBMWatsonTTS(ctx context.Context, apiKey, baseURL, voice, text string) ([]byte, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("ibm watson tts: missing api key")
	}
	v := strings.TrimSpace(voice)
	if v == "" {
		v = "en-US_AllisonV3Voice"
	}
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "https://api.us-south.text-to-speech.watson.cloud.ibm.com"
	}
	base = strings.TrimRight(base, "/")
	reqURL := base + "/v1/synthesize?voice=" + v
	body := `<speak version="1.0"><prosody rate="0%">` + xmlEscapeTTS(text) + `</prosody></speak>`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Accept", "audio/mpeg")
	req.SetBasicAuth("apikey", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ibm watson tts: %s: %s", resp.Status, truncateTTSErr(data))
	}
	return data, nil
}

func xmlEscapeTTS(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

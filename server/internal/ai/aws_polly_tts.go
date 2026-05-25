package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AWSPollyTTS Amazon Polly SynthesizeSpeech（SigV4 简化实现）。
// apiKey=AccessKeyId，config.api_secret=SecretAccessKey；voice_code=VoiceId。
func AWSPollyTTS(ctx context.Context, accessKeyID, secretAccessKey, region, voiceID, text string) ([]byte, error) {
	accessKeyID = strings.TrimSpace(accessKeyID)
	secretAccessKey = strings.TrimSpace(secretAccessKey)
	if accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("aws polly: missing credentials")
	}
	if region == "" {
		region = "us-east-1"
	}
	if voiceID == "" {
		voiceID = "Joanna"
	}
	host := "polly." + region + ".amazonaws.com"
	endpoint := "https://" + host + "/v1/speech"
	params := url.Values{}
	params.Set("Action", "SynthesizeSpeech")
	params.Set("Version", "2016-06-10")
	params.Set("OutputFormat", "mp3")
	params.Set("Text", text)
	params.Set("VoiceId", voiceID)
	params.Set("TextType", "text")
	body := params.Encode()
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	canonicalURI := "/v1/speech"
	canonicalQueryString := ""
	canonicalHeaders := "content-type:application/x-www-form-urlencoded\nhost:" + host + "\nx-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-date"
	payloadHash := sha256HexBytes([]byte(body))
	canonicalRequest := "POST\n" + canonicalURI + "\n" + canonicalQueryString + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHash
	credentialScope := dateStamp + "/" + region + "/polly/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256HexBytes([]byte(canonicalRequest))
	signingKey := aws4SigningKey(secretAccessKey, dateStamp, region, "polly")
	signature := hex.EncodeToString(hmacSHA256Bytes(signingKey, []byte(stringToSign)))
	authorization := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKeyID, credentialScope, signedHeaders, signature,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Authorization", authorization)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("aws polly: %s: %s", resp.Status, truncateTTSErr(data))
	}
	return data, nil
}

func aws4SigningKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256Bytes([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256Bytes(kDate, []byte(region))
	kService := hmacSHA256Bytes(kRegion, []byte(service))
	return hmacSHA256Bytes(kService, []byte("aws4_request"))
}

// VolcengineTTS 火山引擎豆包语音合成（OpenSpeech）。
func VolcengineTTS(ctx context.Context, appID, accessToken, cluster, voiceType, text string) ([]byte, error) {
	appID = strings.TrimSpace(appID)
	accessToken = strings.TrimSpace(accessToken)
	if appID == "" || accessToken == "" {
		return nil, fmt.Errorf("volcengine tts: missing app_id or access_token")
	}
	if cluster == "" {
		cluster = "volcano_tts"
	}
	if voiceType == "" {
		voiceType = "BV001_streaming"
	}
	body, _ := json.Marshal(map[string]any{
		"app": map[string]string{
			"appid":   appID,
			"token":   accessToken,
			"cluster": cluster,
		},
		"user": map[string]string{"uid": "xlangai"},
		"audio": map[string]any{
			"voice_type":   voiceType,
			"encoding":     "mp3",
			"speed_ratio":  1.0,
			"volume_ratio": 1.0,
			"pitch_ratio":  1.0,
		},
		"request": map[string]any{
			"reqid":     fmt.Sprintf("%d", time.Now().UnixNano()),
			"text":      text,
			"text_type": "plain",
			"operation": "query",
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openspeech.bytedance.com/api/v1/tts", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer;"+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("volcengine tts: %s: %s", resp.Status, truncateTTSErr(b))
	}
	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 3000 && parsed.Data == "" {
		return nil, fmt.Errorf("volcengine tts: %d %s", parsed.Code, parsed.Message)
	}
	if parsed.Data == "" {
		return nil, fmt.Errorf("volcengine tts: empty data")
	}
	return base64.StdEncoding.DecodeString(parsed.Data)
}

// PlayHTTTS PlayHT 语音合成（简化 REST）。
func PlayHTTTS(ctx context.Context, apiKey, userID, voiceID, text string) ([]byte, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("playht tts: missing api key")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return nil, fmt.Errorf("playht tts: config user_id required")
	}
	voice := strings.TrimSpace(voiceID)
	if voice == "" {
		return nil, fmt.Errorf("playht tts: empty voice")
	}
	body, _ := json.Marshal(map[string]string{
		"text":         text,
		"voice":        voice,
		"voice_engine": "PlayHT2.0",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.play.ht/api/v2/tts/stream", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("X-User-Id", uid)
	req.Header.Set("Accept", "audio/mpeg")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("playht tts: %s: %s", resp.Status, truncateTTSErr(data))
	}
	return data, nil
}

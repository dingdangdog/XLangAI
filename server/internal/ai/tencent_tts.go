package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// TencentCloudTTS 腾讯云语音合成 TextToVoice。
// voice_code 填 VoiceType 数字（如 1001）；config 可选 codec: mp3。
func TencentCloudTTS(ctx context.Context, secretID, secretKey, region, voiceCode, codec, text string) ([]byte, error) {
	voiceType, err := strconv.Atoi(strings.TrimSpace(voiceCode))
	if err != nil || voiceType <= 0 {
		return nil, fmt.Errorf("tencent tts: voice_code must be numeric VoiceType")
	}
	if codec == "" {
		codec = "mp3"
	}
	payload := map[string]any{
		"Text":      text,
		"SessionId": fmt.Sprintf("xlangai-%d", voiceType),
		"ModelType": 1,
		"VoiceType": voiceType,
		"Codec":     codec,
		"Volume":    0,
		"Speed":     0,
		"ProjectId": 0,
	}
	b, err := TencentCloudTC3(ctx, secretID, secretKey, region, "tts.tencentcloudapi.com", "tts", "TextToVoice", "2019-08-23", payload)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Response struct {
			Audio string `json:"Audio"`
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if parsed.Response.Error != nil {
		return nil, fmt.Errorf("tencent tts: %s", parsed.Response.Error.Message)
	}
	if parsed.Response.Audio == "" {
		return nil, fmt.Errorf("tencent tts: empty audio")
	}
	return base64.StdEncoding.DecodeString(parsed.Response.Audio)
}

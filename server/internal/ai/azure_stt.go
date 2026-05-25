package ai

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

type azureSTTShortResponse struct {
	RecognitionStatus string `json:"RecognitionStatus"`
	DisplayText       string `json:"DisplayText"`
}

// MapLanguageCodeToAzureSTTLocale 将 sys_languages.code（如 en、zh、ja、ko）映射为 Azure 语音识别 locale。
func MapLanguageCodeToAzureSTTLocale(langCode string) string {
	c := strings.ToLower(strings.TrimSpace(langCode))
	if c == "" {
		return "en-US"
	}
	// 粤语 / 香港：须在泛化 zh 之前判断
	if c == "yue" || strings.HasPrefix(c, "yue-") || strings.HasPrefix(c, "zh-hk") || strings.HasPrefix(c, "zh_hk") {
		return "zh-HK"
	}
	if strings.HasPrefix(c, "zh-tw") || strings.HasPrefix(c, "zh_tw") {
		return "zh-TW"
	}
	if i := strings.IndexByte(c, '-'); i > 0 {
		prefix := c[:i]
		switch prefix {
		case "zh", "cmn":
			return "zh-CN"
		case "en":
			return "en-US"
		case "ja":
			return "ja-JP"
		case "ko":
			return "ko-KR"
		case "es":
			return "es-ES"
		case "fr":
			return "fr-FR"
		case "de":
			return "de-DE"
		case "pt":
			return "pt-BR"
		case "it":
			return "it-IT"
		case "ru":
			return "ru-RU"
		case "ar":
			return "ar-SA"
		case "hi":
			return "hi-IN"
		}
	}
	switch c {
	case "zh", "cmn":
		return "zh-CN"
	case "yue":
		return "zh-HK"
	case "en":
		return "en-US"
	case "ja":
		return "ja-JP"
	case "ko":
		return "ko-KR"
	case "es":
		return "es-ES"
	case "fr":
		return "fr-FR"
	case "de":
		return "de-DE"
	case "pt":
		return "pt-BR"
	case "it":
		return "it-IT"
	case "ru":
		return "ru-RU"
	case "ar":
		return "ar-SA"
	case "hi":
		return "hi-IN"
	default:
		return "en-US"
	}
}

// AzureSpeechShortAudioTranscribe 调用 Azure 认知服务短音频 REST（interactive）。
// 音频须为 WAV（PCM 16-bit mono，采样率与 Content-Type 中声明一致，常用 16000Hz）。
// @see https://learn.microsoft.com/azure/ai-services/speech-service/rest-speech-to-text
func AzureSpeechShortAudioTranscribe(ctx context.Context, subscriptionKey, region, locale string, wavPCM16kMono []byte) (string, error) {
	subscriptionKey = strings.TrimSpace(subscriptionKey)
	region = strings.TrimSpace(strings.ToLower(region))
	locale = strings.TrimSpace(locale)
	if locale == "" {
		locale = "en-US"
	}
	if subscriptionKey == "" || region == "" {
		return "", fmt.Errorf("azure stt: missing subscription key or region")
	}
	if len(wavPCM16kMono) == 0 {
		return "", fmt.Errorf("azure stt: empty audio")
	}

	base := fmt.Sprintf("https://%s.stt.speech.microsoft.com/speech/recognition/interactive/cognitiveservices/v1", region)
	q := url.Values{}
	q.Set("language", locale)
	q.Set("format", "simple")
	reqURL := base + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(wavPCM16kMono))
	if err != nil {
		return "", err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", subscriptionKey)
	req.Header.Set("Content-Type", "audio/wav; codecs=audio/pcm; samplerate=16000")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("azure stt api error %d: %s", resp.StatusCode, string(body))
	}

	var r azureSTTShortResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("azure stt: invalid json: %w", err)
	}
	switch strings.TrimSpace(r.RecognitionStatus) {
	case "Success":
		if strings.TrimSpace(r.DisplayText) == "" {
			return "", fmt.Errorf("azure stt: empty display text")
		}
		return strings.TrimSpace(r.DisplayText), nil
	default:
		return "", fmt.Errorf("azure stt: recognition status %q", r.RecognitionStatus)
	}
}

// AzureSTTExtraConfig STT 行 config JSON 中可选字段（region 优先于全局环境变量）。
type AzureSTTExtraConfig struct {
	Region string `json:"region"`
	Locale string `json:"locale"`
}

// ParseAzureSTTConfig 解析 sys_stt_service_configs.config 中的 Azure 扩展字段。
func ParseAzureSTTConfig(configJSON string) AzureSTTExtraConfig {
	var c AzureSTTExtraConfig
	_ = json.Unmarshal([]byte(strings.TrimSpace(configJSON)), &c)
	return c
}

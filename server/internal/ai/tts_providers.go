package ai

import (
	"encoding/json"
	"strings"
)

const (
	ProviderOpenAIRest      = "openai_rest"
	ProviderAzureSpeechREST = "azure_speech_rest"
	ProviderAliyunNLS       = "aliyun_nls"
	ProviderGeminiTTS       = "gemini_tts"
	ProviderTencentTTS      = "tencent_tts"
	ProviderGoogleCloudTTS  = "google_cloud_tts"
	ProviderElevenLabs      = "elevenlabs"
	ProviderBaiduTTS        = "baidu_tts"
	ProviderAWSPolly        = "aws_polly"
	ProviderDeepgramTTS     = "deepgram"
	ProviderMinimaxTTS      = "minimax"
	ProviderIBMWatsonTTS    = "ibm_watson"
	ProviderXunfeiTTS       = "xunfei"
	ProviderVolcengineTTS   = "volcengine"
	ProviderPlayHT          = "playht"
)

// NormalizeTTSProvider 统一后台 provider 取值。
func NormalizeTTSProvider(p string) string {
	switch strings.TrimSpace(strings.ToLower(p)) {
	case "", "openai", "openai_rest":
		return ProviderOpenAIRest
	case "azure", "azure_speech", "azure_speech_rest", "azure_tts", "microsoft":
		return ProviderAzureSpeechREST
	case "aliyun", "aliyun_nls", "alibaba", "alibaba_tts":
		return ProviderAliyunNLS
	case "gemini", "gemini_tts", "google_gemini", "google_gemini_tts":
		return ProviderGeminiTTS
	case "tencent", "tencent_tts", "tencent_cloud":
		return ProviderTencentTTS
	case "google", "google_cloud", "google_cloud_tts", "google_tts":
		return ProviderGoogleCloudTTS
	case "elevenlabs", "eleven_labs":
		return ProviderElevenLabs
	case "baidu", "baidu_tts":
		return ProviderBaiduTTS
	case "aws", "aws_polly", "polly", "amazon_polly":
		return ProviderAWSPolly
	case "deepgram", "deepgram_tts":
		return ProviderDeepgramTTS
	case "minimax", "minimax_tts":
		return ProviderMinimaxTTS
	case "ibm", "ibm_watson", "watson":
		return ProviderIBMWatsonTTS
	case "xunfei", "iflytek", "xfyun":
		return ProviderXunfeiTTS
	case "volcengine", "volcano", "bytedance", "doubao":
		return ProviderVolcengineTTS
	case "playht", "play_ht":
		return ProviderPlayHT
	default:
		return strings.TrimSpace(p)
	}
}

// TTSExtraConfig 从 sys_tts_service_configs.config JSON 解析。
type TTSExtraConfig struct {
	OutputFormat string `json:"output_format"`
	Region       string `json:"region"`
	AppKey       string `json:"app_key"`
	AppID        string `json:"app_id"`
	APISecret    string `json:"api_secret"`
	SecretID     string `json:"secret_id"`
	Format       string `json:"format"`
	SampleRate   int    `json:"sample_rate"`
	LanguageCode string `json:"language_code"`
	AudioEncoding string `json:"audio_encoding"`
	Codec        string `json:"codec"`
	ModelID      string `json:"model_id"`
	Cluster      string `json:"cluster"`
	UserID       string `json:"user_id"`
	Cuid         string `json:"cuid"`
	Spd          int    `json:"spd"`
	Pit          int    `json:"pit"`
	Vol          int    `json:"vol"`
}

func ParseTTSExtraConfig(configJSON string) TTSExtraConfig {
	var ex TTSExtraConfig
	if configJSON != "" {
		_ = json.Unmarshal([]byte(configJSON), &ex)
	}
	return ex
}

// TencentSecretID 解析腾讯云 SecretId（config.secret_id / app_id）。
func (ex TTSExtraConfig) TencentSecretID() string {
	if s := strings.TrimSpace(ex.SecretID); s != "" {
		return s
	}
	return strings.TrimSpace(ex.AppID)
}

// TencentSecretKey 由调用方传入 apiKey；AWS secret 用 APISecret。
func (ex TTSExtraConfig) AWSSecretKey(fallback string) string {
	if s := strings.TrimSpace(ex.APISecret); s != "" {
		return s
	}
	return fallback
}

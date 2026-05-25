package ai

import (
	"context"
	"strings"
)

// TTSResult 合成结果与 MIME（用于选择存储扩展名）。
type TTSResult struct {
	Audio    []byte
	MimeType string
}

// TTSRequest 合成请求（由 handler 从库记录组装）。
type TTSRequest struct {
	Provider     string
	BaseURL      string
	APIKey       string
	Region       string
	ModelCode    string
	ConfigJSON   string
	VoiceCode    string
	Text         string
	AliyunAppKey string
	TencentSecretID string
}

// Synthesize 按 provider 调用对应 TTS。
func Synthesize(ctx context.Context, req TTSRequest) (*TTSResult, error) {
	provider := NormalizeTTSProvider(req.Provider)
	ex := ParseTTSExtraConfig(req.ConfigJSON)

	switch provider {
	case ProviderOpenAIRest:
		model := strings.TrimSpace(req.ModelCode)
		if model == "" || model == "-" {
			model = "tts-1"
		}
		base := req.BaseURL
		if base == "" {
			base = "https://api.openai.com"
		}
		audio, err := OpenAITTSSpeech(ctx, NormalizeOpenAIBaseURL(base), req.APIKey, model, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderAzureSpeechREST:
		outFmt, rJSON := ex.OutputFormat, ex.Region
		if outFmt == "" {
			outFmt, rJSON = ParseAzureOutputFormat(req.ConfigJSON)
		}
		region := strings.TrimSpace(req.Region)
		if region == "" {
			region = rJSON
		}
		audio, err := AzureTTSSpeechREST(ctx, req.APIKey, region, outFmt, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderAliyunNLS:
		appKey := strings.TrimSpace(ex.AppKey)
		if appKey == "" {
			appKey = strings.TrimSpace(req.AliyunAppKey)
		}
		region := strings.TrimSpace(req.Region)
		if region == "" {
			region = strings.TrimSpace(ex.Region)
		}
		audio, mime, err := AliyunNLSTTS(ctx, req.APIKey, appKey, region, req.VoiceCode, ex.Format, ex.SampleRate, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: mime}, nil

	case ProviderGeminiTTS:
		audio, mime, err := GeminiTTSSpeech(ctx, req.APIKey, req.BaseURL, req.ModelCode, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: mime}, nil

	case ProviderTencentTTS:
		sid := strings.TrimSpace(req.TencentSecretID)
		if sid == "" {
			sid = ex.TencentSecretID()
		}
		region := strings.TrimSpace(req.Region)
		if region == "" {
			region = ex.Region
		}
		codec := ex.Codec
		if codec == "" {
			codec = "mp3"
		}
		audio, err := TencentCloudTTS(ctx, sid, req.APIKey, region, req.VoiceCode, codec, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderGoogleCloudTTS:
		enc := ex.AudioEncoding
		if enc == "" {
			enc = "MP3"
		}
		audio, err := GoogleCloudTTSSpeech(ctx, req.APIKey, req.VoiceCode, ex.LanguageCode, enc, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderElevenLabs:
		model := strings.TrimSpace(req.ModelCode)
		if model == "" || model == "-" {
			model = ex.ModelID
		}
		audio, err := ElevenLabsTTS(ctx, req.APIKey, req.BaseURL, req.VoiceCode, model, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderBaiduTTS:
		audio, err := BaiduTTSSpeech(ctx, req.APIKey, ex.Cuid, req.VoiceCode, ex.Spd, ex.Pit, ex.Vol, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderAWSPolly:
		accessKeyID := ex.TencentSecretID()
		if accessKeyID == "" {
			accessKeyID = strings.TrimSpace(ex.AppID)
		}
		if accessKeyID == "" {
			accessKeyID = strings.TrimSpace(req.TencentSecretID)
		}
		secretKey := ex.AWSSecretKey(strings.TrimSpace(req.APIKey))
		region := strings.TrimSpace(req.Region)
		if region == "" {
			region = ex.Region
		}
		audio, err := AWSPollyTTS(ctx, accessKeyID, secretKey, region, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderDeepgramTTS:
		model := strings.TrimSpace(req.ModelCode)
		if model == "" || model == "-" {
			model = ex.ModelID
		}
		audio, err := DeepgramTTS(ctx, req.APIKey, model, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderMinimaxTTS:
		audio, err := MinimaxTTS(ctx, req.APIKey, req.BaseURL, req.ModelCode, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderIBMWatsonTTS:
		audio, err := IBMWatsonTTS(ctx, req.APIKey, req.BaseURL, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderXunfeiTTS:
		appID := strings.TrimSpace(ex.AppID)
		audio, err := XunfeiTTS(ctx, appID, req.APIKey, ex.APISecret, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderVolcengineTTS:
		appID := strings.TrimSpace(ex.AppID)
		cluster := ex.Cluster
		audio, err := VolcengineTTS(ctx, appID, req.APIKey, cluster, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	case ProviderPlayHT:
		audio, err := PlayHTTTS(ctx, req.APIKey, ex.UserID, req.VoiceCode, req.Text)
		if err != nil {
			return nil, err
		}
		return &TTSResult{Audio: audio, MimeType: "audio/mpeg"}, nil

	default:
		return nil, &ErrUnsupportedTTSProvider{Provider: req.Provider}
	}
}

// ErrUnsupportedTTSProvider 未实现的 TTS 实现类型。
type ErrUnsupportedTTSProvider struct {
	Provider string
}

func (e *ErrUnsupportedTTSProvider) Error() string {
	return "unsupported tts provider: " + e.Provider
}

// MimeToAudioExt 根据 MIME 选择存储扩展名。
func MimeToAudioExt(mime string) string {
	m := strings.ToLower(strings.TrimSpace(mime))
	switch {
	case strings.Contains(m, "wav"):
		return ".wav"
	case strings.Contains(m, "ogg"):
		return ".ogg"
	default:
		return ".mp3"
	}
}

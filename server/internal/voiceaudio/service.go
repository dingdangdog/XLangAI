package voiceaudio

import (
	"context"
	"strings"

	"xlangai/server/internal/ai"
	"xlangai/server/internal/llmchat"
	"xlangai/server/internal/repository"
)

// Input 多模态语音轮入参。
type Input struct {
	Protocol      string
	BaseURL       string
	APIKey        string
	ModelCode     string
	Config        string
	VoiceCode     string
	RoleConfig    *string
	SystemPrompt  string
	Messages      []llmchat.Message
	UserAudio     []byte
	UserAudioMime string
	UserText      string
	SynthesisType string
}

// Output 生成结果。
type Output struct {
	Text      string
	Audio     []byte
	AudioMime string
	Usage     *ai.ChatUsage
}

// Generate 按角色合成策略生成助手回复。
func Generate(ctx context.Context, in Input) (*Output, error) {
	st := repository.NormalizeSynthesisType(in.SynthesisType)
	protocol := strings.TrimSpace(strings.ToLower(in.Protocol))
	if protocol == "" {
		protocol = "openai"
	}

	switch st {
	case repository.SynthesisNativeAudioInText:
		if len(in.UserAudio) > 0 {
			return generateAudioInTextOut(ctx, in, protocol)
		}
		text, usage, err := llmchat.Chat(
			ctx,
			llmchat.ServiceInputFromRepo(protocol, in.BaseURL, in.APIKey, in.ModelCode, in.Config),
			in.SystemPrompt,
			append(in.Messages, llmchat.Message{Role: "user", Content: in.UserText}),
		)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(text) == "" {
			return nil, ErrNoText
		}
		return &Output{Text: text, Usage: usage}, nil

	case repository.SynthesisNativeAudioIO:
		return generateAudioIO(ctx, in, protocol)

	default:
		return nil, ErrConfig
	}
}

func generateAudioInTextOut(ctx context.Context, in Input, protocol string) (*Output, error) {
	switch protocol {
	case "gemini", "google_gemini":
		out, err := geminiGenerate(ctx, in, false)
		if err != nil {
			return nil, err
		}
		if out.Text == "" {
			return nil, ErrNoText
		}
		return out, nil
	default:
		return nil, ErrUnsupported
	}
}

func generateAudioIO(ctx context.Context, in Input, protocol string) (*Output, error) {
	switch protocol {
	case "gemini", "google_gemini":
		out, err := geminiGenerate(ctx, in, true)
		if err != nil {
			return nil, err
		}
		if out.Text == "" {
			return nil, ErrNoTranscript
		}
		if len(out.Audio) == 0 {
			return nil, ErrNoAudio
		}
		return out, nil
	default:
		return nil, ErrUnsupported
	}
}

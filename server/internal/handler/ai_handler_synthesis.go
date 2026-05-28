package handler

import (
	"context"
	"errors"
	"log"
	"strings"

	"xlangai/server/config"
	"xlangai/server/internal/ai"
	"xlangai/server/internal/entity"
	"xlangai/server/internal/llmchat"
	"xlangai/server/internal/model"
	"xlangai/server/internal/repository"
	"xlangai/server/internal/voiceaudio"

	"github.com/gin-gonic/gin"
)

var (
	errVoiceSynthConfig      = errors.New("voice role synthesis config invalid")
	errVoiceSynthFailed      = errors.New("voice synthesis failed")
	errVoiceSynthNoTranscript = errors.New("voice synthesis: no transcript")
	errVoiceSynthNoAudio     = errors.New("voice synthesis: no audio")
	errTTSFailed             = errors.New("tts synthesis failed")
)

func (h *AIHandler) loadVoiceRoleForConv(ctx context.Context, conv *model.Conversation) (*entity.VoiceRole, string, error) {
	if conv.VoiceRoleID == nil || strings.TrimSpace(*conv.VoiceRoleID) == "" {
		return nil, repository.SynthesisTTS, nil
	}
	vr, err := h.voice.GetEntityByID(ctx, strings.TrimSpace(*conv.VoiceRoleID))
	if err != nil || vr == nil {
		return nil, "", errVoiceSynthConfig
	}
	return vr, repository.NormalizeSynthesisType(vr.SynthesisType), nil
}

func (h *AIHandler) resolveRoleLLM(ctx context.Context, vr *entity.VoiceRole) (*repository.LLMServiceConfig, error) {
	if vr == nil || vr.LlmServiceConfigID == nil {
		return nil, errVoiceSynthConfig
	}
	id := strings.TrimSpace(*vr.LlmServiceConfigID)
	if id == "" {
		return nil, errVoiceSynthConfig
	}
	cfg, err := h.llmCfg.GetByID(ctx, id)
	if err != nil || cfg == nil {
		return nil, errVoiceSynthConfig
	}
	if h.llmAPIKey(cfg) == "" {
		return nil, errLLMKeyMissing
	}
	return cfg, nil
}

func messagesBeforeUser(msgs []*model.Message, userMsgID string) []llmchat.Message {
	out := make([]llmchat.Message, 0, len(msgs))
	for _, m := range msgs {
		if m.ID == userMsgID {
			continue
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		out = append(out, llmchat.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

func (h *AIHandler) runNativeSynthesisTurn(
	c *gin.Context,
	ctx context.Context,
	conv *model.Conversation,
	userID string,
	userMsg *model.Message,
	vr *entity.VoiceRole,
	synthType string,
	useTTS *bool,
	userAudio []byte,
	userAudioMime string,
	userText string,
	systemPrompt string,
) (reply string, audioURL *string, usage *ai.ChatUsage, llmCfgID string, err error) {
	if synthType == repository.SynthesisNativeAudioIO {
		if useTTS != nil && !*useTTS {
			return "", nil, nil, "", errVoiceSynthConfig
		}
	}

	llmCfg, err := h.resolveRoleLLM(ctx, vr)
	if err != nil {
		return "", nil, nil, "", err
	}
	llmCfgID = llmCfg.ID
	apiKey := h.llmAPIKey(llmCfg)

	msgs, err := h.msg.ListByConversation(ctx, conv.ID, 20, nil)
	if err != nil {
		return "", nil, nil, "", err
	}
	history := messagesBeforeUser(msgs, userMsg.ID)

	out, err := voiceaudio.Generate(ctx, voiceaudio.Input{
		Protocol:      llmCfg.Protocol,
		BaseURL:       llmCfg.BaseURL,
		APIKey:        apiKey,
		ModelCode:     llmCfg.ModelCode,
		Config:        llmCfg.Config,
		VoiceCode:     vr.VoiceCode,
		RoleConfig:    vr.Config,
		SystemPrompt:  systemPrompt,
		Messages:      history,
		UserAudio:     userAudio,
		UserAudioMime: userAudioMime,
		UserText:      strings.TrimSpace(userText),
		SynthesisType: synthType,
	})
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[voiceaudio] synth_type=%s: %v", synthType, err)
		}
		switch {
		case errors.Is(err, voiceaudio.ErrNoTranscript):
			return "", nil, nil, "", errVoiceSynthNoTranscript
		case errors.Is(err, voiceaudio.ErrNoAudio):
			return "", nil, nil, "", errVoiceSynthNoAudio
		case errors.Is(err, voiceaudio.ErrConfig), errors.Is(err, voiceaudio.ErrProviderNotReady):
			return "", nil, nil, "", errVoiceSynthConfig
		case errors.Is(err, voiceaudio.ErrUnsupported):
			return "", nil, nil, "", llmchat.ErrUnsupportedProtocol
		default:
			return "", nil, nil, "", errVoiceSynthFailed
		}
	}

	reply = ai.NormalizeAssistantReply(out.Text)
	usage = out.Usage

	if synthType == repository.SynthesisNativeAudioIO && len(out.Audio) > 0 {
		ext := ai.MimeToAudioExt(out.AudioMime)
		audioURL, err = h.saveAssistantTTS(ctx, out.Audio, ext)
		if err != nil {
			return "", nil, nil, "", err
		}
		if h.cfg.TTSLoudnessNorm {
			// optional loudness on native audio — skip for minimal scope
		}
	}

	return reply, audioURL, usage, llmCfgID, nil
}

func userContentForNativeVoice(content string) string {
	c := strings.TrimSpace(content)
	if c != "" {
		return c
	}
	return "（语音输入）"
}

func (h *AIHandler) applyNativeLoudness(ctx context.Context, cfg *config.Config, audio []byte, mime string) ([]byte, string) {
	if !cfg.TTSLoudnessNorm {
		return audio, mime
	}
	opts := &ai.TTSLoudnessOptions{TargetLUFS: cfg.TTSTargetLUFS}
	norm, normMime, normErr := ai.NormalizeTTSLoudness(ctx, cfg.FFmpegPath, audio, mime, opts)
	if normErr == nil && len(norm) > 0 {
		if normMime != "" {
			return norm, normMime
		}
		return norm, mime
	}
	return audio, mime
}

package handler

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"unicode/utf8"

	"xlangai/server/internal/ai"
	"xlangai/server/internal/model"
	"xlangai/server/internal/repository"
)

// ScenarioOpeningEnsurer 幂等写入场景开场助手消息（预设模板 + TTS，非 LLM）。
type ScenarioOpeningEnsurer interface {
	EnsureScenarioOpening(ctx context.Context, conv *model.Conversation, userID string) error
}

var openingConvLocks sync.Map

func mergeOpeningTemplate(template, voiceName string) string {
	name := strings.TrimSpace(voiceName)
	if name == "" {
		name = "AI"
	}
	out := strings.TrimSpace(template)
	out = strings.ReplaceAll(out, "{{voice_role_name}}", name)
	out = strings.ReplaceAll(out, "{{name}}", name)
	out = strings.ReplaceAll(out, "{name}", name)
	return out
}

func (h *AIHandler) EnsureScenarioOpening(ctx context.Context, conv *model.Conversation, userID string) error {
	if conv == nil || h.opening == nil {
		return nil
	}
	scenarioCode := strings.TrimSpace(conv.ScenarioCode)
	if scenarioCode == "" || scenarioCode == "free" {
		return nil
	}

	lockIface, _ := openingConvLocks.LoadOrStore(conv.ID, &sync.Mutex{})
	lock := lockIface.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	count, err := h.msg.CountByConversation(ctx, conv.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	langCode := strings.TrimSpace(conv.LanguageCode)
	if langCode == "" && h.lang != nil {
		if code, lerr := h.lang.GetCodeByID(ctx, conv.LanguageID); lerr == nil {
			langCode = strings.TrimSpace(code)
		}
	}

	tpl, err := h.opening.ResolveTemplate(ctx, scenarioCode, langCode)
	if err != nil || strings.TrimSpace(tpl) == "" {
		if h.cfg.VerboseLogs {
			log.Printf("[opening] no template scenario=%s lang=%s: %v", scenarioCode, langCode, err)
		}
		return err
	}

	voiceName := ""
	if conv.VoiceRoleID != nil && strings.TrimSpace(*conv.VoiceRoleID) != "" {
		if vr, verr := h.voice.GetByID(ctx, strings.TrimSpace(*conv.VoiceRoleID)); verr == nil && vr != nil {
			voiceName = strings.TrimSpace(vr.Name)
		}
	}
	content := mergeOpeningTemplate(tpl, voiceName)
	if content == "" {
		return nil
	}

	useTTS := true
	audioURL, synthErr := h.synthesizeAssistantText(ctx, userID, conv, content, &useTTS)
	if synthErr != nil && h.cfg.VerboseLogs {
		log.Printf("[opening] TTS skipped conv=%s: %v", conv.ID, synthErr)
		audioURL = nil
	}

	// 并发请求可能在 TTS 期间已写入消息，再次校验。
	count, err = h.msg.CountByConversation(ctx, conv.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err = h.msg.Create(ctx, repository.CreateMessageInput{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        content,
		AudioURL:       audioURL,
	})
	return err
}

func (h *AIHandler) synthesizeAssistantText(ctx context.Context, userID string, conv *model.Conversation, text string, useTTS *bool) (*string, error) {
	enabled := useTTS == nil || *useTTS
	voiceID := conv.VoiceRoleID
	if !enabled || voiceID == nil || *voiceID == "" {
		return nil, nil
	}

	vrEntity, err := h.voice.GetEntityByID(ctx, *voiceID)
	if err != nil || vrEntity == nil {
		return nil, nil
	}
	if repository.NormalizeSynthesisType(vrEntity.SynthesisType) != repository.SynthesisTTS {
		return nil, nil
	}
	if vrEntity.TtsServiceConfigID == nil || strings.TrimSpace(*vrEntity.TtsServiceConfigID) == "" || strings.TrimSpace(vrEntity.VoiceCode) == "" {
		return nil, errTTSFailed
	}

	ttsRow, err := h.ttsCfg.GetByID(ctx, strings.TrimSpace(*vrEntity.TtsServiceConfigID))
	if err != nil || ttsRow == nil {
		return nil, errTTSFailed
	}

	apiKey, region, tBase := h.resolveTTSCredentials(ttsRow)
	ex := ai.ParseTTSExtraConfig(ttsRow.ConfigJSON)

	result, err := ai.Synthesize(ctx, ai.TTSRequest{
		Provider:        ttsRow.Provider,
		BaseURL:         tBase,
		APIKey:          apiKey,
		Region:          region,
		ModelCode:       ttsRow.ModelCode,
		ConfigJSON:      ttsRow.ConfigJSON,
		VoiceCode:       vrEntity.VoiceCode,
		Text:            text,
		AliyunAppKey:    strings.TrimSpace(ex.AppKey),
		TencentSecretID: ex.TencentSecretID(),
	})
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[tts] synthesize provider=%s code=%q: %v", ttsRow.Provider, ttsRow.Code, err)
		}
		return nil, errTTSFailed
	}
	if result == nil || len(result.Audio) == 0 {
		return nil, errTTSFailed
	}
	if h.cfg.TTSLoudnessNorm {
		opts := &ai.TTSLoudnessOptions{TargetLUFS: h.cfg.TTSTargetLUFS}
		norm, normMime, normErr := ai.NormalizeTTSLoudness(ctx, h.cfg.FFmpegPath, result.Audio, result.MimeType, opts)
		if normErr == nil && len(norm) > 0 {
			result.Audio = norm
			if normMime != "" {
				result.MimeType = normMime
			}
		} else {
			boost, boostMime, boostErr := ai.BoostTTSVolumeFallback(ctx, h.cfg.FFmpegPath, result.Audio, result.MimeType)
			if boostErr == nil && len(boost) > 0 {
				result.Audio = boost
				if boostMime != "" {
					result.MimeType = boostMime
				}
			} else if errors.Is(normErr, ai.ErrFFmpegNotFound) || errors.Is(boostErr, ai.ErrFFmpegNotFound) {
				log.Printf("[tts] WARNING: ffmpeg not found — TTS audio is not loudness-normalized. Install ffmpeg or set XLANGAI_FFMPEG_PATH")
			} else if h.cfg.VerboseLogs && normErr != nil {
				log.Printf("[tts] loudness normalize skipped: %v", normErr)
			}
		}
	}
	if h.usage != nil && userID != "" {
		charCount := utf8.RuneCountInString(text)
		_ = h.usage.Record(ctx, repository.UsageRecord{
			UserID:          userID,
			TTSCalls:        1,
			TTSChars:        charCount,
			ServiceType:     repository.ServiceUsageTTS,
			ServiceConfigID: ttsRow.ID,
			ServiceRequests: 1,
			ServiceUnits:    int64(charCount),
		})
	}
	ext := ai.MimeToAudioExt(result.MimeType)
	return h.saveAssistantTTS(ctx, result.Audio, ext)
}

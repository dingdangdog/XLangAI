package handler

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"xlangai/server/config"
	"xlangai/server/internal/ai"
	"xlangai/server/internal/authz"
	"xlangai/server/internal/llmchat"
	"xlangai/server/internal/media"
	"xlangai/server/internal/messagemeta"
	"xlangai/server/internal/model"
	"xlangai/server/internal/objectstore"
	"xlangai/server/internal/repository"
	"xlangai/server/internal/translate"

	"github.com/gin-gonic/gin"
)

const openAIBaseURL = "https://api.openai.com"

type AIHandler struct {
	cfg          *config.Config
	msg          *repository.MessageRepo
	sys          *repository.SystemRepo
	conv         *repository.ConvRepo
	lang         *repository.LangRepo
	llmCfg       *repository.LLMConfigRepo
	sttCfg       *repository.STTConfigRepo
	ttsCfg       *repository.TtsConfigRepo
	translateCfg *repository.TranslateConfigRepo
	voice        *repository.VoiceRepo
	usage        *repository.UsageRepo
	az           *authz.Service
	media        *media.Service
}

func NewAIHandler(
	cfg *config.Config,
	msg *repository.MessageRepo,
	sys *repository.SystemRepo,
	conv *repository.ConvRepo,
	lang *repository.LangRepo,
	llmCfg *repository.LLMConfigRepo,
	sttCfg *repository.STTConfigRepo,
	ttsCfg *repository.TtsConfigRepo,
	translateCfg *repository.TranslateConfigRepo,
	voice *repository.VoiceRepo,
	usage *repository.UsageRepo,
	az *authz.Service,
	mediaSvc *media.Service,
) *AIHandler {
	return &AIHandler{
		cfg: cfg, msg: msg, sys: sys, conv: conv, lang: lang,
		llmCfg: llmCfg, sttCfg: sttCfg, ttsCfg: ttsCfg, translateCfg: translateCfg,
		voice: voice, usage: usage, az: az, media: mediaSvc,
	}
}

func (h *AIHandler) resolveTranslateInput(ctx context.Context) (translate.ServiceInput, error) {
	cfg, err := h.translateCfg.GetActive(ctx)
	if err != nil || cfg == nil {
		return translate.ServiceInput{}, translate.ErrProviderNotReady
	}
	protocol := strings.TrimSpace(cfg.Protocol)
	if protocol == "" {
		protocol = "openai"
	}
	in := translate.ServiceInput{
		Protocol:    protocol,
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		APISecret:   cfg.APISecret,
		ModelCode:   cfg.ModelCode,
		LlmConfigID: cfg.LlmConfigID,
		Config:      translate.ParseProviderConfig(cfg.Config),
	}
	if strings.EqualFold(protocol, "openai") {
		llmID := strings.TrimSpace(cfg.LlmConfigID)
		if llmID != "" {
			llm, err := h.llmCfg.GetByID(ctx, llmID)
			if err != nil || llm == nil || !llmchat.IsOpenAICompatible(llm.Protocol) {
				return translate.ServiceInput{}, translate.ErrProviderNotReady
			}
			if k := h.llmAPIKey(llm); k != "" {
				in.APIKey = k
			}
			if bu := strings.TrimSpace(llm.BaseURL); bu != "" {
				in.BaseURL = bu
			}
			if mc := strings.TrimSpace(llm.ModelCode); mc != "" {
				in.ModelCode = mc
			}
		}
	}
	return in, nil
}

// Translate 将文本翻译为客户端指定的区域/语言（BCP 47），使用 sys_translate_service_configs 中唯一启用的服务商。
func (h *AIHandler) Translate(c *gin.Context) {
	uid := strings.TrimSpace(c.GetString("user_id"))
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Text         string `json:"text" binding:"required"`
		TargetLocale string `json:"target_locale" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is empty"})
		return
	}
	target := strings.TrimSpace(req.TargetLocale)
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_locale is required"})
		return
	}

	if h.az != nil {
		if p := CtxPrincipal(c); p != nil {
			if err := h.az.EnsureChatQuota(c.Request.Context(), p); err != nil {
				h.writeQuotaError(c, err)
				return
			}
		}
	}

	in, err := h.resolveTranslateInput(c.Request.Context())
	if err != nil {
		if errors.Is(err, translate.ErrProviderNotReady) {
			payload := gin.H{"error": "翻译服务未就绪", "code": "TRANSLATE_PROVIDER_NOT_READY"}
			if h.cfg.VerboseLogs {
				payload["detail"] = err.Error()
			}
			c.JSON(http.StatusServiceUnavailable, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	translateCfg, _ := h.translateCfg.GetActive(c.Request.Context())
	maxChars := in.Config.MaxChars
	if maxChars <= 0 {
		maxChars = 12000
	}
	if utf8.RuneCountInString(text) > maxChars {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text too long"})
		return
	}

	out, err := translate.Translate(c.Request.Context(), in, text, target)
	if err != nil {
		if errors.Is(err, translate.ErrProviderNotReady) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "翻译服务未就绪", "code": "TRANSLATE_PROVIDER_NOT_READY"})
			return
		}
		if errors.Is(err, translate.ErrUnsupportedProtocol) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "不支持的翻译协议", "code": "TRANSLATE_PROTOCOL_UNSUPPORTED"})
			return
		}
		log.Printf("[translate] user_id=%s protocol=%s: %v", uid, in.Protocol, err)
		payload := gin.H{"error": "翻译失败", "code": "TRANSLATE_FAILED"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusInternalServerError, payload)
		return
	}
	if h.usage != nil && translateCfg != nil {
		charCount := utf8.RuneCountInString(text)
		_ = h.usage.Record(c.Request.Context(), repository.UsageRecord{
			UserID:          uid,
			TranslateCalls:  1,
			TranslateChars:  charCount,
			ServiceType:     repository.ServiceUsageTranslate,
			ServiceConfigID: translateCfg.ID,
			ServiceRequests: 1,
			ServiceUnits:    int64(charCount),
		})
	}
	c.JSON(http.StatusOK, gin.H{"text": out})
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req struct {
		Content          string `json:"content" binding:"required"`
		UseTTS           *bool  `json:"use_tts"`
		OriginalAudioURL string `json:"original_audio_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, userID, ok := h.loadAuthorizedConversation(c)
	if !ok {
		return
	}

	var originalAudioURL *string
	if u := strings.TrimSpace(req.OriginalAudioURL); u != "" {
		if err := h.media.ValidateObjectURL(c.Request.Context(), media.ScopeUserRecording, u); err != nil {
			if errors.Is(err, media.ErrInvalidObjectURL) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid original_audio_url"})
				return
			}
			writeMediaStorageError(c, err)
			return
		}
		originalAudioURL = &u
	}

	resp, err := h.runConversationTurn(c, conv, userID, req.Content, req.UseTTS, originalAudioURL, nil)
	if err != nil {
		if resp != nil {
			if skipped, _ := resp["ai_skipped"].(bool); skipped {
				log.Printf("[llm] user_id=%s conv_id=%s: %v", userID, conv.ID, err)
				c.JSON(http.StatusOK, resp)
				return
			}
		}
		h.writeConversationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AIHandler) VoiceChat(c *gin.Context) {
	conv, userID, ok := h.loadAuthorizedConversation(c)
	if !ok {
		return
	}

	useTTS := parseOptionalBool(c.PostForm("use_tts"))
	ctx := c.Request.Context()

	audioURLForm := strings.TrimSpace(c.PostForm("audio_url"))
	var audioBytes []byte
	var filename string
	var audioURL *string

	if audioURLForm != "" {
		if err := h.media.ValidateObjectURL(ctx, media.ScopeUserRecording, audioURLForm); err != nil {
			if errors.Is(err, media.ErrInvalidObjectURL) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio_url"})
				return
			}
			writeMediaStorageError(c, err)
			return
		}
		data, err := h.media.DownloadObjectBytes(ctx, audioURLForm)
		if err != nil {
			if errors.Is(err, objectstore.ErrNotConfigured) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "对象存储未配置或凭证不完整，请在管理后台配置并启用"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch uploaded audio"})
			return
		}
		audioBytes = data
		filename = "voice-input" + normalizedAudioExt(audioURLForm)
		audioURL = &audioURLForm
	} else {
		file, err := c.FormFile("audio")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "audio file or audio_url is required"})
			return
		}
		if file.Size <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is empty"})
			return
		}
		if file.Size > 25<<20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "audio file too large"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read audio file"})
			return
		}
		defer src.Close()

		audioBytes, err = io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read audio bytes"})
			return
		}

		filename = strings.TrimSpace(file.Filename)
		if filename == "" {
			filename = "voice-input.webm"
		}

		savedURL, err := h.saveUserRecording(ctx, audioBytes, normalizedAudioExt(filename))
		if err != nil {
			if errors.Is(err, media.ErrClientOnlyStorage) {
				c.JSON(http.StatusForbidden, gin.H{"error": "user recording storage is client-only"})
				return
			}
			if errors.Is(err, objectstore.ErrNotConfigured) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "对象存储未配置或凭证不完整，请在管理后台配置并启用"})
				return
			}
			if h.cfg.VerboseLogs {
				log.Printf("VoiceChat: saveBinaryAudio conv_id=%q: %v", c.Param("id"), err)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save audio"})
			return
		}
		audioURL = savedURL
	}

	transcript := strings.TrimSpace(c.PostForm("transcript"))
	var text string
	if transcript != "" {
		text = transcript
	} else if h.chatQuotaExceeded(c) {
		text = ""
	} else {
		var err error
		text, err = h.transcribeAudio(c, conv, userID, filename, audioBytes)
		if err != nil {
			h.writeConversationError(c, err)
			return
		}
	}

	var sttPtr *string
	if text != "" {
		sttPtr = &text
	}
	resp, err := h.runConversationTurn(c, conv, userID, text, useTTS, audioURL, sttPtr)
	if err != nil {
		if resp != nil {
			if skipped, _ := resp["ai_skipped"].(bool); skipped {
				log.Printf("[llm] user_id=%s conv_id=%s: %v", userID, conv.ID, err)
				resp["transcript"] = text
				c.JSON(http.StatusOK, resp)
				return
			}
		}
		h.writeConversationError(c, err)
		return
	}
	resp["transcript"] = text
	c.JSON(http.StatusOK, resp)
}

func (h *AIHandler) runConversationTurn(
	c *gin.Context,
	conv *model.Conversation,
	userID string,
	userContent string,
	useTTS *bool,
	originalAudioURL *string,
	sttText *string,
) (gin.H, error) {
	ctx := c.Request.Context()

	userMsg, err := h.msg.Create(ctx, repository.CreateMessageInput{
		ConversationID:   conv.ID,
		Role:             "user",
		Content:          userContent,
		OriginalAudioURL: originalAudioURL,
		STTText:          sttText,
	})
	if err != nil {
		return nil, err
	}

	if quotaErr := h.checkChatQuota(c); quotaErr != nil {
		updated, uerr := h.msg.UpdateMetadata(ctx, userMsg.ID, messagemeta.MarshalStatus(messagemeta.StatusQuotaExceeded))
		if uerr == nil && updated != nil {
			userMsg = updated
		}
		return gin.H{
			"user_message": userMsg,
			"message":      nil,
			"ai_skipped":   true,
			"code":         authz.QuotaErrorCode(quotaErr),
		}, nil
	}

	systemPrompt, err := h.sys.ResolveSystemPromptForConversation(ctx, conv.LanguageID, conv.PromptID, conv.VoiceRoleID)
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}

	llmCfg, err := h.resolveLLMForConversation(ctx, conv)
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}
	apiKey := h.llmAPIKey(llmCfg)
	if apiKey == "" {
		return h.failUserTurn(ctx, userMsg, errLLMKeyMissing)
	}

	msgs, err := h.msg.ListByConversation(ctx, conv.ID, 20, nil)
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}

	chatMsgs := make([]struct{ Role, Content string }, 0, len(msgs))
	for _, m := range msgs {
		chatMsgs = append(chatMsgs, struct{ Role, Content string }{Role: m.Role, Content: m.Content})
	}

	llmMsgs := make([]llmchat.Message, len(chatMsgs))
	for i, m := range chatMsgs {
		llmMsgs[i] = llmchat.Message{Role: m.Role, Content: m.Content}
	}
	llmIn := llmchat.ServiceInputFromRepo(
		llmCfg.Protocol, llmCfg.BaseURL, apiKey, llmCfg.ModelCode, llmCfg.Config,
	)
	debitTokens := false
	if h.az != nil {
		if p := CtxPrincipal(c); p != nil {
			var err error
			debitTokens, err = h.az.UsesTokenWalletForNextTurn(ctx, p)
			if err != nil {
				return h.failUserTurn(ctx, userMsg, err)
			}
		}
	}
	reply, usage, err := llmchat.Chat(ctx, llmIn, strings.TrimSpace(systemPrompt), llmMsgs)
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}
	reply = ai.NormalizeAssistantReply(reply)

	audioURL, err := h.synthesizeReply(c, userID, conv, reply, useTTS)
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}

	assist, err := h.msg.Create(ctx, repository.CreateMessageInput{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        reply,
		AudioURL:       audioURL,
	})
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}

	if updated, uerr := h.msg.UpdateMetadata(ctx, userMsg.ID, messagemeta.MarshalStatus(messagemeta.StatusSuccess)); uerr == nil && updated != nil {
		userMsg = updated
	}

	if h.usage != nil {
		rec := repository.UsageRecord{
			UserID:    userID,
			ChatTurns: 1,
		}
		if usage != nil && usage.TotalTokens > 0 {
			rec.LLMTokens = usage.TotalTokens
			rec.ServiceType = repository.ServiceUsageLLM
			rec.ServiceConfigID = llmCfg.ID
			rec.ServiceRequests = 1
			rec.ServiceUnits = int64(usage.TotalTokens)
		}
		_ = h.usage.Record(ctx, rec)
	}

	if h.az != nil {
		if debitTokens && usage != nil && usage.TotalTokens > 0 {
			_ = h.az.DeductChatTokens(ctx, userID, int64(usage.TotalTokens))
		}
	}

	return gin.H{
		"user_message": userMsg,
		"message":      assist,
	}, nil
}

func (h *AIHandler) failUserTurn(ctx context.Context, userMsg *model.Message, cause error) (gin.H, error) {
	if userMsg != nil {
		if updated, err := h.msg.UpdateMetadata(ctx, userMsg.ID, messagemeta.MarshalStatus(messagemeta.StatusFailed)); err == nil && updated != nil {
			userMsg = updated
		}
	}
	return gin.H{
		"user_message": userMsg,
		"message":      nil,
		"ai_skipped":   true,
		"code":         "AI_FAILED",
	}, cause
}

func (h *AIHandler) checkChatQuota(c *gin.Context) error {
	if h.az == nil {
		return nil
	}
	p := CtxPrincipal(c)
	if p == nil {
		return nil
	}
	return h.az.EnsureChatQuota(c.Request.Context(), p)
}

func (h *AIHandler) chatQuotaExceeded(c *gin.Context) bool {
	return h.checkChatQuota(c) != nil
}

func (h *AIHandler) writeQuotaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, authz.ErrQuotaDaily):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "QUOTA_DAILY"})
	case errors.Is(err, authz.ErrQuotaMonthly):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "QUOTA_MONTHLY"})
	case errors.Is(err, authz.ErrQuotaTokens):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "QUOTA_TOKENS"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *AIHandler) synthesizeReply(c *gin.Context, userID string, conv *model.Conversation, reply string, useTTS *bool) (*string, error) {
	enabled := useTTS == nil || *useTTS
	voiceID := conv.VoiceRoleID
	if !enabled || voiceID == nil || *voiceID == "" {
		return nil, nil
	}

	vr, err := h.voice.GetByID(c.Request.Context(), *voiceID)
	if err != nil || vr == nil || vr.TtsServiceConfigID == "" || vr.VoiceCode == "" {
		return nil, nil
	}

	ttsRow, err := h.ttsCfg.GetByID(c.Request.Context(), vr.TtsServiceConfigID)
	if err != nil || ttsRow == nil {
		return nil, nil
	}

	apiKey, region, tBase := h.resolveTTSCredentials(ttsRow)
	ex := ai.ParseTTSExtraConfig(ttsRow.ConfigJSON)

	result, err := ai.Synthesize(c.Request.Context(), ai.TTSRequest{
		Provider:        ttsRow.Provider,
		BaseURL:         tBase,
		APIKey:          apiKey,
		Region:          region,
		ModelCode:       ttsRow.ModelCode,
		ConfigJSON:      ttsRow.ConfigJSON,
		VoiceCode:       vr.VoiceCode,
		Text:            reply,
		AliyunAppKey:    strings.TrimSpace(ex.AppKey),
		TencentSecretID: ex.TencentSecretID(),
	})
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[tts] synthesize provider=%s code=%q: %v", ttsRow.Provider, ttsRow.Code, err)
		}
		return nil, nil
	}
	if result == nil || len(result.Audio) == 0 {
		return nil, nil
	}
	if h.cfg.TTSLoudnessNorm {
		opts := &ai.TTSLoudnessOptions{TargetLUFS: h.cfg.TTSTargetLUFS}
		norm, normMime, normErr := ai.NormalizeTTSLoudness(c.Request.Context(), h.cfg.FFmpegPath, result.Audio, result.MimeType, opts)
		if normErr == nil && len(norm) > 0 {
			result.Audio = norm
			if normMime != "" {
				result.MimeType = normMime
			}
		} else {
			boost, boostMime, boostErr := ai.BoostTTSVolumeFallback(c.Request.Context(), h.cfg.FFmpegPath, result.Audio, result.MimeType)
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
		charCount := utf8.RuneCountInString(reply)
		_ = h.usage.Record(c.Request.Context(), repository.UsageRecord{
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
	return h.saveAssistantTTS(c.Request.Context(), result.Audio, ext)
}

func (h *AIHandler) resolveTTSCredentials(ttsRow *repository.TtsServiceConfig) (apiKey, region, baseURL string) {
	if ttsRow == nil {
		return "", "", ""
	}
	ex := ai.ParseTTSExtraConfig(ttsRow.ConfigJSON)
	apiKey = strings.TrimSpace(ttsRow.APIKey)
	region = strings.TrimSpace(ttsRow.Region)
	if region == "" {
		region = strings.TrimSpace(ex.Region)
	}
	baseURL = strings.TrimSpace(ttsRow.BaseURL)
	if region == "" {
		if _, rj := ai.ParseAzureOutputFormat(ttsRow.ConfigJSON); rj != "" {
			region = rj
		}
	}
	return apiKey, region, baseURL
}

func (h *AIHandler) transcribeAzureSpeech(ctx context.Context, conv *model.Conversation, stt *repository.STTServiceConfig, audioBytes []byte) (string, error) {
	apiKey := strings.TrimSpace(stt.APIKey)
	if apiKey == "" {
		return "", errAzureSTTMissingKey
	}
	extra := ai.ParseAzureSTTConfig(stt.Config)
	region := strings.TrimSpace(extra.Region)
	if region == "" {
		return "", errAzureSTTRegionMissing
	}

	locale := strings.TrimSpace(extra.Locale)
	mode := strings.TrimSpace(strings.ToLower(h.cfg.STTLanguageMode))
	if mode == "" {
		mode = "auto"
	}
	if locale == "" && mode == "target" && h.lang != nil {
		if code, err := h.lang.GetCodeByID(ctx, conv.LanguageID); err == nil {
			locale = ai.MapLanguageCodeToAzureSTTLocale(code)
		}
	}
	if locale == "" {
		// auto 与未配置扩展 locale 时：按会话目标语言选择 Azure locale（支持中英日韩等）。多语言混杂自动切分需 Speech SDK 连续识别，此处不展开。
		if h.lang != nil {
			if code, err := h.lang.GetCodeByID(ctx, conv.LanguageID); err == nil {
				locale = ai.MapLanguageCodeToAzureSTTLocale(code)
			}
		}
	}
	if locale == "" {
		locale = "en-US"
	}

	if h.cfg.VerboseLogs {
		log.Printf("[stt] azure region=%q locale=%q bytes_in=%d", region, locale, len(audioBytes))
	}

	wav, err := ai.TranscodeToAzureSTTWavPCM16k(ctx, h.cfg.FFmpegPath, audioBytes)
	if err != nil {
		return "", err
	}
	text, err := ai.AzureSpeechShortAudioTranscribe(ctx, apiKey, region, locale, wav)
	if err != nil && h.cfg.VerboseLogs {
		log.Printf("[stt] azure transcribe failed: %v", err)
	}
	return text, err
}

func (h *AIHandler) transcribeAudio(c *gin.Context, conv *model.Conversation, userID, filename string, audioBytes []byte) (string, error) {
	ctx := c.Request.Context()
	sttCfg, sttErr := h.sttCfg.GetFirstActive(ctx)
	if sttErr == nil && strings.TrimSpace(strings.ToLower(sttCfg.Protocol)) == ai.ProviderAzureSpeechREST {
		text, err := h.transcribeAzureSpeech(ctx, conv, sttCfg, audioBytes)
		if err == nil && h.usage != nil && userID != "" {
			h.recordSTTUsage(ctx, userID, sttCfg.ID, audioBytes)
		}
		return text, err
	}

	apiKey := ""
	baseURL := openAIBaseURL
	modelCode := "whisper-1"

	if sttErr != nil || sttCfg == nil {
		return "", errSTTNotConfigured
	}
	apiKey = strings.TrimSpace(sttCfg.APIKey)
	if apiKey == "" {
		return "", errSTTKeyMissing
	}
	if u := strings.TrimSpace(sttCfg.BaseURL); u != "" {
		baseURL = u
	}
	if m := strings.TrimSpace(sttCfg.ModelCode); m != "" {
		modelCode = m
	}

	normForSTT := ai.NormalizeOpenAIBaseURL(baseURL)
	if strings.Contains(strings.ToLower(normForSTT), "integrate.api.nvidia.com") {
		return "", errSTTNvidiaIntegrateNeedsCompatSTT
	}

	rawBase := baseURL
	normBase := ai.NormalizeOpenAIBaseURL(baseURL)
	if h.cfg.VerboseLogs {
		src := "sys_stt_service_configs"
		sttCode := sttCfg.Code
		log.Printf("[stt] source=%s code=%q raw_base_url=%q normalized_base=%q model=%q file=%q bytes=%d",
			src, sttCode, rawBase, normBase, modelCode, filename, len(audioBytes))
		if rawBase != normBase {
			log.Printf("[stt] 提示: Base URL 已自动去掉末尾 /v1；后台「AI 服务配置」里 Base URL 应填根地址（如 https://api.openai.com），不要带 /v1")
		}
	}

	var whisperLang *string
	mode := strings.TrimSpace(strings.ToLower(h.cfg.STTLanguageMode))
	if mode == "" {
		mode = "auto"
	}
	if mode == "target" && h.lang != nil {
		code, err := h.lang.GetCodeByID(ctx, conv.LanguageID)
		if err == nil {
			whisperLang = ai.WhisperLanguageHintFromSysLanguageCode(code)
		}
	}

	client := ai.NewOpenAIClient(apiKey, "", baseURL)
	text, trErr := client.Transcribe(ctx, filename, audioBytes, modelCode, whisperLang)
	if trErr == nil && h.usage != nil && userID != "" && sttCfg != nil {
		h.recordSTTUsage(ctx, userID, sttCfg.ID, audioBytes)
	}
	if trErr != nil && h.cfg.VerboseLogs {
		log.Printf("[stt] transcribe failed: %v", trErr)
	}
	return text, trErr
}

func (h *AIHandler) recordSTTUsage(ctx context.Context, userID, configID string, audioBytes []byte) {
	if h.usage == nil || userID == "" || configID == "" {
		return
	}
	_ = h.usage.Record(ctx, repository.UsageRecord{
		UserID:          userID,
		STTCalls:        1,
		STTAudioBytes:   int64(len(audioBytes)),
		ServiceType:     repository.ServiceUsageSTT,
		ServiceConfigID: configID,
		ServiceRequests: 1,
		ServiceUnits:    int64(len(audioBytes)),
	})
}

func (h *AIHandler) loadAuthorizedConversation(c *gin.Context) (*model.Conversation, string, bool) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, "", false
	}

	convID := c.Param("id")
	conv, err := h.conv.GetByID(c.Request.Context(), convID)
	if err != nil || conv == nil || conv.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return nil, "", false
	}

	p := CtxPrincipal(c)
	if p != nil {
		if err := p.EnsureLanguageAllowed(conv.LanguageID); err != nil {
			if errors.Is(err, authz.ErrForbiddenLanguage) {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN_LANGUAGE"})
				return nil, "", false
			}
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return nil, "", false
		}
		if err := p.EnsureVoiceAllowed(conv.VoiceRoleID); err != nil {
			if errors.Is(err, authz.ErrForbiddenVoice) {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "FORBIDDEN_VOICE"})
				return nil, "", false
			}
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return nil, "", false
		}
	}

	return conv, userID, true
}

var errLLMKeyMissing = errors.New("llm api key not configured")

var (
	errAzureSTTMissingKey    = errors.New("azure stt: api key not configured")
	errAzureSTTRegionMissing = errors.New("azure stt: region not configured")
	errSTTNotConfigured      = errors.New("stt: no active service config")
	errSTTKeyMissing         = errors.New("stt: api key not configured")
)

// errSTTNvidiaIntegrateNeedsCompatSTT：integrate.api.nvidia.com 不提供 /v1/audio/transcriptions；须在后台 STT 配置中填写实际 ASR 服务根地址。
var errSTTNvidiaIntegrateNeedsCompatSTT = errors.New("stt: integrate.api.nvidia.com has no OpenAI-compatible /v1/audio/transcriptions; configure sys_stt_service_configs with a host that implements POST /v1/audio/transcriptions")

func (h *AIHandler) llmAPIKey(cfg *repository.LLMServiceConfig) string {
	if cfg == nil {
		return ""
	}
	return strings.TrimSpace(cfg.APIKey)
}

// resolveLLMForConversation 优先使用会话上的 llm_config_id（JSON 仍为 ai_config_id）；若该行无可用密钥或已失效，则回退到 sort_order 最前的活跃 LLM 配置。
func (h *AIHandler) resolveLLMForConversation(ctx context.Context, conv *model.Conversation) (*repository.LLMServiceConfig, error) {
	seen := make(map[string]struct{})
	var ids []string
	if conv.LLMConfigID != nil {
		if id := strings.TrimSpace(*conv.LLMConfigID); id != "" {
			ids = append(ids, id)
			seen[id] = struct{}{}
		}
	}
	if def, err := h.llmCfg.GetFirstActive(ctx); err == nil && def != nil {
		if _, ok := seen[def.ID]; !ok {
			ids = append(ids, def.ID)
		}
	}
	for _, id := range ids {
		cfg, err := h.llmCfg.GetByID(ctx, id)
		if err != nil || cfg == nil {
			continue
		}
		if h.llmAPIKey(cfg) == "" {
			continue
		}
		return cfg, nil
	}
	return nil, errLLMKeyMissing
}

func (h *AIHandler) writeConversationError(c *gin.Context, err error) {
	log.Printf("ai conversation %s %s conv_id=%q: %v", c.Request.Method, c.FullPath(), c.Param("id"), err)
	switch {
	case errors.Is(err, errLLMKeyMissing):
		payload := gin.H{"error": "大模型服务未配置，请在管理后台配置 LLM 服务", "code": "LLM_KEY_MISSING"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusServiceUnavailable, payload)
	case errors.Is(err, llmchat.ErrUnsupportedProtocol):
		payload := gin.H{"error": "不支持的 LLM 协议，请在管理后台调整该配置的 protocol", "code": "LLM_PROTOCOL_UNSUPPORTED"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusNotImplemented, payload)
	case errors.Is(err, llmchat.ErrProviderNotReady):
		payload := gin.H{"error": "LLM 服务未配置密钥", "code": "LLM_PROVIDER_NOT_READY"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusServiceUnavailable, payload)
	case errors.Is(err, errSTTNvidiaIntegrateNeedsCompatSTT):
		payload := gin.H{
			"error": "STT 配置的 Base URL 不能为 NVIDIA integrate 网关（无语音转写 REST）。请在管理后台 STT 服务配置中填写实际 ASR 厂商根地址与 API Key",
			"code":  "STT_NVDA_INTEGRATE",
		}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusServiceUnavailable, payload)
	case errors.Is(err, errSTTNotConfigured):
		{
			payload := gin.H{"error": "语音转写未配置，请在管理后台添加并启用 STT 服务配置", "code": "STT_NOT_CONFIGURED"}
			if h.cfg.VerboseLogs {
				payload["detail"] = err.Error()
			}
			c.JSON(http.StatusServiceUnavailable, payload)
		}
	case errors.Is(err, errSTTKeyMissing):
		{
			payload := gin.H{"error": "语音转写未配置 API Key，请在管理后台 STT 服务配置中填写", "code": "STT_KEY_MISSING"}
			if h.cfg.VerboseLogs {
				payload["detail"] = err.Error()
			}
			c.JSON(http.StatusServiceUnavailable, payload)
		}
	case errors.Is(err, errAzureSTTMissingKey):
		payload := gin.H{"error": "Azure 语音转写未配置密钥，请在管理后台 STT 服务配置中填写 API Key", "code": "AZURE_STT_KEY_MISSING"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusServiceUnavailable, payload)
	case errors.Is(err, errAzureSTTRegionMissing):
		payload := gin.H{"error": "Azure 语音转写未配置区域：请在 STT 扩展 JSON 或区域字段中填写 region", "code": "AZURE_STT_REGION_MISSING"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusServiceUnavailable, payload)
	case errors.Is(err, ai.ErrFFmpegNotFound):
		payload := gin.H{"error": "服务器需要 ffmpeg 将录音转为 Azure 所需格式。请安装 ffmpeg 并加入 PATH，或设置 XLANGAI_FFMPEG_PATH", "code": "FFMPEG_NOT_FOUND"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusServiceUnavailable, payload)
	default:
		payload := gin.H{"error": "服务异常，请稍后再试"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusInternalServerError, payload)
	}
}

func (h *AIHandler) saveUserRecording(ctx context.Context, audioBytes []byte, ext string) (*string, error) {
	res, err := h.media.SaveUserRecording(ctx, audioBytes, ext, audioMimeType("x"+ext))
	if err != nil {
		return nil, err
	}
	u := res.URL
	return &u, nil
}

func (h *AIHandler) saveAssistantTTS(ctx context.Context, audioBytes []byte, ext string) (*string, error) {
	res, err := h.media.SaveAssistantTTS(ctx, audioBytes, ext, audioMimeType("x"+ext))
	if err != nil {
		return nil, err
	}
	u := res.URL
	return &u, nil
}

func (h *AIHandler) ServeAudio(c *gin.Context) {
	name := c.Param("filename")
	if name == "" || len(name) > 100 || strings.ContainsAny(name, "/\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}
	data, err := h.media.ReadAudio(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Data(http.StatusOK, audioMimeType(name), data)
}

func parseOptionalBool(raw string) *bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		v := true
		return &v
	case "0", "false", "no", "off":
		v := false
		return &v
	default:
		return nil
	}
}

func normalizedAudioExt(filename string) string {
	switch ext := strings.ToLower(strings.TrimSpace(filepath.Ext(filename))); ext {
	case ".mp3", ".wav", ".m4a", ".webm", ".mp4", ".mpeg", ".mpga", ".ogg", ".oga":
		return ext
	default:
		return ".webm"
	}
}

func audioMimeType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".wav":
		return "audio/wav"
	case ".m4a", ".mp4":
		return "audio/mp4"
	case ".webm":
		return "audio/webm"
	case ".ogg", ".oga":
		return "audio/ogg"
	default:
		return "audio/mpeg"
	}
}

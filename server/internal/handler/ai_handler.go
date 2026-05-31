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
	opening      *repository.ScenarioOpeningRepo
	userRepo     *repository.UserRepo
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
	opening *repository.ScenarioOpeningRepo,
	userRepo *repository.UserRepo,
	usage *repository.UsageRepo,
	az *authz.Service,
	mediaSvc *media.Service,
) *AIHandler {
	return &AIHandler{
		cfg: cfg, msg: msg, sys: sys, conv: conv, lang: lang,
		llmCfg: llmCfg, sttCfg: sttCfg, ttsCfg: ttsCfg, translateCfg: translateCfg,
		voice: voice, opening: opening, userRepo: userRepo, usage: usage, az: az, media: mediaSvc,
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

	resp, err := h.runConversationTurn(c, conv, userID, req.Content, req.UseTTS, originalAudioURL, nil, nil, "")
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

	_, synthType, vrErr := h.loadVoiceRoleForConv(ctx, conv)
	if vrErr != nil {
		h.writeConversationError(c, vrErr)
		return
	}
	nativeVoice := repository.IsNativeSynthesis(synthType) && len(audioBytes) > 0

	var text string
	var sttPtr *string
	if nativeVoice {
		text = strings.TrimSpace(userContentForNativeVoice(""))
	} else {
		transcript := strings.TrimSpace(c.PostForm("transcript"))
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
		if text != "" {
			sttPtr = &text
		}
	}

	audioMime := audioMimeType(filename)
	resp, err := h.runConversationTurn(c, conv, userID, text, useTTS, audioURL, sttPtr, audioBytes, audioMime)
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
	userAudio []byte,
	userAudioMime string,
) (gin.H, error) {
	ctx := c.Request.Context()

	vrEntity, synthType, vrErr := h.loadVoiceRoleForConv(ctx, conv)
	if vrErr != nil {
		// 尚未创建 userMsg，直接返回错误
		return nil, vrErr
	}
	if repository.IsNativeSynthesis(synthType) {
		userContent = userContentForNativeVoice(userContent)
	}

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

	systemPrompt, err := h.sys.ResolveSystemPromptForConversation(ctx, conv.LanguageID, conv.PromptID, conv.VoiceRoleID, scenarioCodePtr(conv.ScenarioCode))
	if err != nil {
		return h.failUserTurn(ctx, userMsg, err)
	}

	debitTokens := false
	if h.az != nil {
		if p := CtxPrincipal(c); p != nil {
			debitTokens, err = h.az.UsesTokenWalletForNextTurn(ctx, p)
			if err != nil {
				return h.failUserTurn(ctx, userMsg, err)
			}
		}
	}

	var reply string
	var audioURL *string
	var usage *ai.ChatUsage
	var llmCfgID string

	if repository.IsNativeSynthesis(synthType) && vrEntity != nil {
		genUserText := userContent
		if len(userAudio) > 0 {
			genUserText = ""
		}
		reply, audioURL, usage, llmCfgID, err = h.runNativeSynthesisTurn(
			c, ctx, conv, userID, userMsg, vrEntity, synthType, useTTS,
			userAudio, userAudioMime, genUserText, systemPrompt,
		)
	} else {
		llmCfg, llmErr := h.resolveLLMForConversation(ctx, conv)
		if llmErr != nil {
			return h.failUserTurn(ctx, userMsg, llmErr)
		}
		apiKey := h.llmAPIKey(llmCfg)
		if apiKey == "" {
			return h.failUserTurn(ctx, userMsg, errLLMKeyMissing)
		}
		llmCfgID = llmCfg.ID

		msgs, listErr := h.msg.ListByConversation(ctx, conv.ID, 20, nil)
		if listErr != nil {
			return h.failUserTurn(ctx, userMsg, listErr)
		}
		llmMsgs := make([]llmchat.Message, 0, len(msgs))
		for _, m := range msgs {
			if strings.TrimSpace(m.Content) == "" {
				continue
			}
			llmMsgs = append(llmMsgs, llmchat.Message{Role: m.Role, Content: m.Content})
		}
		llmIn := llmchat.ServiceInputFromRepo(
			llmCfg.Protocol, llmCfg.BaseURL, apiKey, llmCfg.ModelCode, llmCfg.Config,
		)
		reply, usage, err = llmchat.Chat(ctx, llmIn, strings.TrimSpace(systemPrompt), llmMsgs)
		if err != nil {
			return h.failUserTurn(ctx, userMsg, err)
		}
		reply = ai.NormalizeAssistantReply(reply)
		audioURL, err = h.synthesizeReply(c, userID, conv, reply, useTTS)
	}

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
			rec.ServiceConfigID = llmCfgID
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
	return h.synthesizeAssistantText(c.Request.Context(), userID, conv, reply, useTTS)
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

// resolveLLMForConversation 解析 TTS 模式对话 LLM：会话 llm_config_id → 用户默认 LLM → 全局首个 active；native_audio 不走此函数。
func (h *AIHandler) resolveLLMForConversation(ctx context.Context, conv *model.Conversation) (*repository.LLMServiceConfig, error) {
	convLLM := ""
	if conv.LLMConfigID != nil {
		convLLM = strings.TrimSpace(*conv.LLMConfigID)
	}
	userLLM := ""
	if h.userRepo != nil && strings.TrimSpace(conv.UserID) != "" {
		if id, err := h.userRepo.ResolveActiveDefaultLlmConfigID(ctx, conv.UserID); err == nil && id != nil {
			userLLM = *id
		}
	}
	globalLLM := ""
	if def, err := h.llmCfg.GetFirstActive(ctx); err == nil && def != nil {
		globalLLM = def.ID
	}
	ids := repository.OrderedLLMConfigIDsForTTS(convLLM, userLLM, globalLLM)
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
	case errors.Is(err, errVoiceSynthConfig):
		payload := gin.H{"error": "语音角色合成配置无效，请在管理后台检查合成类型与 LLM/TTS 绑定", "code": "VOICE_SYNTH_CONFIG"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusBadRequest, payload)
	case errors.Is(err, errVoiceSynthFailed):
		payload := gin.H{"error": "多模态语音生成失败", "code": "VOICE_SYNTH_FAILED"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusBadGateway, payload)
	case errors.Is(err, errVoiceSynthNoTranscript):
		payload := gin.H{"error": "多模态语音未返回文本", "code": "VOICE_SYNTH_NO_TRANSCRIPT"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusBadGateway, payload)
	case errors.Is(err, errVoiceSynthNoAudio):
		payload := gin.H{"error": "多模态语音未返回音频", "code": "VOICE_SYNTH_NO_AUDIO"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusBadGateway, payload)
	case errors.Is(err, errTTSFailed):
		payload := gin.H{"error": "TTS 合成失败", "code": "TTS_FAILED"}
		if h.cfg.VerboseLogs {
			payload["detail"] = err.Error()
		}
		c.JSON(http.StatusBadGateway, payload)
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

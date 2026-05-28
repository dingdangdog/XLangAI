package repository

import (
	"context"
	"errors"
	"strings"

	"xlangai/server/internal/entity"

	"gorm.io/gorm"
)

type SystemRepo struct {
	db *gorm.DB
}

func NewSystemRepo(db *gorm.DB) *SystemRepo {
	return &SystemRepo{db: db}
}

type Defaults struct {
	LLMConfigID string // sys_llm_service_configs.id
	VoiceID     string // sys_voice_roles.id
	VoiceCode   string // OpenAI voice 如 alloy
	VoiceName   string // 该语言下默认 active 语音角色的展示名（sys_voice_roles.name）
	TTSConfigID string // sys_tts_service_configs.id（与默认语音角色关联，供客户端展示等）
	PromptID    string
	PromptTpl   string
	LangName    string
}

func (r *SystemRepo) GetDefaults(ctx context.Context, langID string) (*Defaults, error) {
	var lang entity.Language
	if err := r.db.WithContext(ctx).Where("id = ?", langID).First(&lang).Error; err != nil {
		return nil, err
	}
	var llm entity.SysLlmServiceConfig
	if err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		First(&llm).Error; err != nil {
		return nil, err
	}
	var prompt entity.PromptTemplate
	if err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", "lang_practice", "active").
		First(&prompt).Error; err != nil {
		return nil, err
	}
	d := &Defaults{
		LLMConfigID: llm.ID,
		PromptID:    prompt.ID,
		PromptTpl:   prompt.Content,
		LangName:    lang.Name,
	}
	var voice entity.VoiceRole
	err := r.db.WithContext(ctx).
		Where("language_id = ? AND status = ?", langID, "active").
		Order("sort_order ASC, created_at ASC").
		First(&voice).Error
	if err == nil {
		d.VoiceID = voice.ID
		d.VoiceCode = voice.VoiceCode
		d.VoiceName = strings.TrimSpace(voice.Name)
		if voice.TtsServiceConfigID != nil {
			d.TTSConfigID = *voice.TtsServiceConfigID
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return d, nil
}

// formatVoiceRolePromptForInjection 将角色专属提示词格式化为可注入系统提示的段落；空内容返回空字符串。
func formatVoiceRolePromptForInjection(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	return "\n\n[Character-specific identity & style]\n" + s
}

// ResolveSystemPromptForConversation 解析会话系统提示：优先会话 prompt_id 对应模板，否则使用语言默认（lang_practice）；{{target_lang}} 替换为语言名称；{{voice_role_name}} 替换为当前会话 TTS 声线展示名（无则语言默认声线名，再否则「语言教练」）；{{voice_role_prompt}} 替换为当前语音角色的 role_prompt（无则空）。
func (r *SystemRepo) ResolveSystemPromptForConversation(ctx context.Context, langID string, promptID *string, voiceRoleID *string) (string, error) {
	def, err := r.GetDefaults(ctx, langID)
	if err != nil {
		return "", err
	}
	tpl := def.PromptTpl
	langName := def.LangName

	if promptID != nil {
		pid := strings.TrimSpace(*promptID)
		if pid != "" {
			var custom entity.PromptTemplate
			err := r.db.WithContext(ctx).
				Where("id = ? AND status = ?", pid, "active").
				First(&custom).Error
			if err == nil && strings.TrimSpace(custom.Content) != "" {
				tpl = custom.Content
			}
		}
	}

	voiceDisplay := strings.TrimSpace(def.VoiceName)
	voiceRolePromptRaw := ""
	if voiceRoleID != nil {
		vid := strings.TrimSpace(*voiceRoleID)
		if vid != "" {
			var vr entity.VoiceRole
			err := r.db.WithContext(ctx).
				Where("id = ? AND status = ?", vid, "active").
				First(&vr).Error
			if err == nil {
				vn := strings.TrimSpace(vr.Name)
				if vn != "" {
					voiceDisplay = vn
				}
				if vr.RolePrompt != nil {
					voiceRolePromptRaw = strings.TrimSpace(*vr.RolePrompt)
				}
			}
		}
	}
	if voiceDisplay == "" {
		voiceDisplay = "语言教练"
	}

	voiceRolePromptRaw = strings.ReplaceAll(voiceRolePromptRaw, "{{target_lang}}", langName)
	voiceRolePromptRaw = strings.ReplaceAll(voiceRolePromptRaw, "{{voice_role_name}}", voiceDisplay)
	voiceRolePrompt := formatVoiceRolePromptForInjection(voiceRolePromptRaw)

	out := strings.ReplaceAll(tpl, "{{target_lang}}", langName)
	out = strings.ReplaceAll(out, "{{voice_role_name}}", voiceDisplay)
	out = strings.ReplaceAll(out, "{{voice_role_prompt}}", voiceRolePrompt)
	return appendPromptFactualContext(out), nil
}

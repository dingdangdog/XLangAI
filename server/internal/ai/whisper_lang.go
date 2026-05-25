package ai

import "strings"

// WhisperLanguageHintFromSysLanguageCode 将会话语言 code（sys_languages.code，可为 BCP-47 前缀）
// 转为 OpenAI Whisper `language` 表单字段可用的 ISO 639-1 提示；无法安全映射时返回 nil（不传 language，走自动识别）。
func WhisperLanguageHintFromSysLanguageCode(langCode string) *string {
	c := strings.ToLower(strings.TrimSpace(langCode))
	if c == "" {
		return nil
	}
	if i := strings.IndexByte(c, '-'); i > 0 {
		c = c[:i]
	}
	switch c {
	case "yue", "cmn":
		s := "zh"
		return &s
	default:
		if len(c) == 2 {
			s := c
			return &s
		}
		return nil
	}
}

package ai

import (
	"regexp"
	"strings"
)

// 覆盖常见 emoji / 绘文字块（含 ✅ U+2705 等），用于 TTS 与展示前兜底清理。
var emojiStripper = regexp.MustCompile(
	`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]` +
		`|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA70}-\x{1FAFF}]` +
		`|[\x{231A}-\x{231B}]|[\x{23E9}-\x{23F3}]|[\x{23F8}-\x{23FA}]`,
)

// NormalizeAssistantReply 去掉 emoji、压缩多余空行并 trim，供入库与 TTS 共用。
func NormalizeAssistantReply(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = emojiStripper.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(s)
}

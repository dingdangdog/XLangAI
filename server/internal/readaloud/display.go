package readaloud

import (
	"strings"

	"xlangai/server/internal/entity"
)

// ResolveCategoryDisplay 按语言 code 解析展示名（仅兼容旧 name_en / 默认 name）。
func ResolveCategoryDisplay(cat *entity.ReadAloudCategory, langCode string) (displayName, displayDesc string) {
	if cat == nil {
		return "", ""
	}
	displayName = strings.TrimSpace(cat.Name)
	if cat.Description != nil {
		displayDesc = strings.TrimSpace(*cat.Description)
	}
	if strings.ToLower(strings.TrimSpace(langCode)) == "en" {
		if cat.NameEn != nil {
			if n := strings.TrimSpace(*cat.NameEn); n != "" {
				displayName = n
			}
		}
		if cat.DescriptionEn != nil {
			if d := strings.TrimSpace(*cat.DescriptionEn); d != "" {
				displayDesc = d
			}
		}
	}
	return displayName, displayDesc
}

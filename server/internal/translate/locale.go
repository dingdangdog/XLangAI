package translate

import "strings"

// TargetForProvider 将 BCP-47 / IETF locale 转为各翻译厂商常用目标码。
func TargetForProvider(protocol, targetLocale string) string {
	tag := strings.TrimSpace(targetLocale)
	if tag == "" {
		return "en"
	}
	primary := tag
	if i := strings.IndexByte(tag, '-'); i > 0 {
		primary = tag[:i]
	}
	primary = strings.ToLower(primary)

	switch protocol {
	case "deepl":
		return deeplTarget(primary, tag)
	case "azure_translator":
		return azureTarget(primary, tag)
	case "google_translate", "baidu_translate", "tencent_translate", "aliyun_translate":
		return primary
	default:
		return tag
	}
}

func deeplTarget(primary, full string) string {
	switch primary {
	case "zh":
		if strings.Contains(strings.ToLower(full), "tw") || strings.Contains(strings.ToLower(full), "hk") {
			return "ZH-HANT"
		}
		return "ZH"
	case "en":
		return "EN-US"
	case "pt":
		return "PT-BR"
	default:
		return strings.ToUpper(primary)
	}
}

func azureTarget(primary, full string) string {
	lower := strings.ToLower(full)
	switch primary {
	case "zh":
		if strings.Contains(lower, "tw") {
			return "zh-Hant"
		}
		if strings.Contains(lower, "hk") || strings.Contains(lower, "yue") {
			return "yue"
		}
		return "zh-Hans"
	case "en":
		return "en"
	default:
		if len(full) >= 2 {
			return full
		}
		return primary
	}
}

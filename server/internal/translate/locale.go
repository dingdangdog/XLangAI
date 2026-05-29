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
	case "youdao_translate":
		return youdaoTarget(primary, tag)
	case "papago_translate":
		return papagoTarget(primary, tag)
	case "xunfei_translate":
		return xunfeiTarget(primary)
	case "google_translate", "baidu_translate", "tencent_translate", "aliyun_translate", "aws_translate", "volcengine_translate", "libretranslate", "ibm_watson_translate":
		return primary
	default:
		return tag
	}
}

func youdaoTarget(primary, full string) string {
	lower := strings.ToLower(full)
	switch primary {
	case "zh":
		if strings.Contains(lower, "tw") || strings.Contains(lower, "hk") || strings.Contains(lower, "hant") {
			return "zh-CHT"
		}
		return "zh-CHS"
	default:
		return primary
	}
}

func papagoTarget(primary, full string) string {
	lower := strings.ToLower(full)
	switch primary {
	case "zh":
		if strings.Contains(lower, "tw") || strings.Contains(lower, "hant") {
			return "zh-TW"
		}
		return "zh-CN"
	case "en":
		return "en"
	case "ja":
		return "ja"
	case "ko":
		return "ko"
	default:
		if len(full) >= 2 {
			return full
		}
		return primary
	}
}

func xunfeiTarget(primary string) string {
	switch primary {
	case "zh":
		return "cn"
	case "en":
		return "en"
	case "ja":
		return "ja"
	case "ko":
		return "ko"
	case "fr":
		return "fr"
	case "es":
		return "es"
	case "ru":
		return "ru"
	case "de":
		return "de"
	default:
		return primary
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

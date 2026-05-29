package translate

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrProviderNotReady = errors.New("translate provider not configured")
	ErrUnsupportedProtocol = errors.New("unsupported translate protocol")
)

func truncateErr(b []byte, max int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

// Translate 按 protocol 路由到对应翻译实现。
func Translate(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	protocol := strings.TrimSpace(strings.ToLower(in.Protocol))
	if protocol == "" {
		protocol = "openai"
	}
	switch protocol {
	case "openai":
		return translateOpenAI(ctx, in, text, targetLocale)
	case "azure_translator":
		return translateAzure(ctx, in, text, targetLocale)
	case "deepl":
		return translateDeepL(ctx, in, text, targetLocale)
	case "google_translate":
		return translateGoogle(ctx, in, text, targetLocale)
	case "baidu_translate":
		return translateBaidu(ctx, in, text, targetLocale)
	case "tencent_translate":
		return translateTencent(ctx, in, text, targetLocale)
	case "aliyun_translate":
		return translateAliyun(ctx, in, text, targetLocale)
	case "aws_translate":
		return translateAWS(ctx, in, text, targetLocale)
	case "youdao_translate":
		return translateYoudao(ctx, in, text, targetLocale)
	case "papago_translate":
		return translatePapago(ctx, in, text, targetLocale)
	case "ibm_watson_translate":
		return translateIBM(ctx, in, text, targetLocale)
	case "libretranslate":
		return translateLibre(ctx, in, text, targetLocale)
	case "xunfei_translate":
		return translateXunfei(ctx, in, text, targetLocale)
	case "volcengine_translate":
		return translateVolcengine(ctx, in, text, targetLocale)
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedProtocol, protocol)
	}
}

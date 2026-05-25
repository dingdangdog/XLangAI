package config

import (
	"bufio"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// prismaDBURLQueryKeys：Prisma DATABASE_URL 查询参数，libpq/pgx 不能作为连接启动参数。
var prismaDBURLQueryKeys = []string{
	"schema",
	"connection_limit",
	"pool_timeout",
	"connect_timeout",
	"socket_timeout",
	"pgbouncer",
	"statement_cache_size",
}

// normalizeDBURL 去掉 Prisma 专用查询参数，供 GORM/pgx 连接 Postgres。
func normalizeDBURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	for _, k := range prismaDBURLQueryKeys {
		q.Del(k)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

type Config struct {
	Port              string
	DBURL             string
	JWTSecret         string
	AudioDir          string
	AvatarDir         string
	RedisURL          string
	LangCacheTTL      time.Duration
	PrincipalCacheTTL time.Duration
	// VerboseLogs：为 true 时部分 5xx 响应带 detail、控制台打印 STT/列表消息等上下文。XLANGAI_VERBOSE_LOGS 显式开关；未设时非 release 默认开启。
	VerboseLogs bool
	// STTLanguageMode：auto 不向 Whisper 传 language（自动识别，适合中英混杂）；target 按会话目标语种传 ISO 639-1 提示（XLANGAI_STT_LANGUAGE_MODE）。
	STTLanguageMode string
	// FFmpegPath：Azure STT 等需将上传音频转 WAV 时使用；默认可执行文件名为 ffmpeg（XLANGAI_FFMPEG_PATH）。
	FFmpegPath string
	// TTSLoudnessNorm：为 true 时在保存助手 TTS 前用 ffmpeg loudnorm 归一化响度（XLANGAI_TTS_LOUDNESS_NORM，默认开启）。
	TTSLoudnessNorm bool
	// TTSTargetLUFS：loudnorm 目标综合响度 LUFS（XLANGAI_TTS_TARGET_LUFS，默认 -16）。
	TTSTargetLUFS float64
	// Apple / Google 内购校验（可选；未配置时 /billing/verify 对应平台返回 503）
	AppleIssuerID                string
	AppleKeyID                   string
	AppleBundleID                string
	ApplePrivateKeyPath          string
	AppleEnvironment             string // sandbox | production
	GooglePlayPackageName        string
	GooglePlayServiceAccountJSON string // 服务账号 JSON 文件路径或内联 JSON
	// GoogleOAuthClientIDs：校验 Google id_token 时允许的 OAuth Client ID（逗号分隔，可含 Android / iOS / Web）。
	GoogleOAuthClientIDs []string
	// AppleSignInClientIDs：校验 Apple identityToken 时允许的 aud（逗号分隔，一般为 App Bundle ID 与/或 Services ID）。
	AppleSignInClientIDs []string
}

func Load() *Config {
	loadLocalEnv()

	port := strings.TrimSpace(os.Getenv("XLANGAI_SERVER_PORT"))
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xlangai?sslmode=disable"
	}
	dbURL = normalizeDBURL(dbURL)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	audioDir := os.Getenv("AUDIO_DIR")
	if audioDir == "" {
		audioDir = "./storage/audio"
	}
	avatarDir := strings.TrimSpace(os.Getenv("AVATAR_DIR"))
	if avatarDir == "" {
		avatarDir = "./storage/avatars"
	}
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	langTTL := 5 * time.Minute
	if v := os.Getenv("LANG_CACHE_TTL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			langTTL = time.Duration(n) * time.Second
		}
	}
	principalTTL := 60 * time.Second
	if v := os.Getenv("PRINCIPAL_CACHE_TTL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			principalTTL = time.Duration(n) * time.Second
		}
	}
	appleBundle := strings.TrimSpace(os.Getenv("XLANGAI_APPLE_BUNDLE_ID"))
	if appleBundle == "" {
		appleBundle = "com.xlangai.ios"
	}
	appleEnv := strings.ToLower(strings.TrimSpace(os.Getenv("XLANGAI_APPLE_ENV")))
	if appleEnv == "" {
		appleEnv = "sandbox"
	}
	gpPkg := strings.TrimSpace(os.Getenv("XLANGAI_GOOGLE_PLAY_PACKAGE"))
	if gpPkg == "" {
		gpPkg = "com.xlangai.android"
	}
	googleAuds := splitComma(os.Getenv("XLANGAI_GOOGLE_OAUTH_CLIENT_IDS"))
	appleAuds := splitComma(os.Getenv("XLANGAI_APPLE_SIGN_IN_CLIENT_IDS"))
	if len(appleAuds) == 0 && appleBundle != "" {
		appleAuds = []string{appleBundle}
	}
	return &Config{
		Port:                         port,
		DBURL:                        dbURL,
		JWTSecret:                    jwtSecret,
		AudioDir:                     audioDir,
		AvatarDir:                    avatarDir,
		RedisURL:                     redisURL,
		LangCacheTTL:                 langTTL,
		PrincipalCacheTTL:            principalTTL,
		VerboseLogs:                  parseVerboseLogs(),
		STTLanguageMode:              strings.TrimSpace(strings.ToLower(os.Getenv("XLANGAI_STT_LANGUAGE_MODE"))),
		FFmpegPath:                   strings.TrimSpace(os.Getenv("XLANGAI_FFMPEG_PATH")),
		TTSLoudnessNorm:              parseTTSLoudnessNorm(),
		TTSTargetLUFS:                parseTTSTargetLUFS(),
		AppleIssuerID:                strings.TrimSpace(os.Getenv("XLANGAI_APPLE_ISSUER_ID")),
		AppleKeyID:                   strings.TrimSpace(os.Getenv("XLANGAI_APPLE_KEY_ID")),
		AppleBundleID:                appleBundle,
		ApplePrivateKeyPath:          strings.TrimSpace(os.Getenv("XLANGAI_APPLE_PRIVATE_KEY_PATH")),
		AppleEnvironment:             appleEnv,
		GooglePlayPackageName:        gpPkg,
		GooglePlayServiceAccountJSON: strings.TrimSpace(os.Getenv("XLANGAI_GOOGLE_PLAY_SERVICE_ACCOUNT_JSON")),
		GoogleOAuthClientIDs:         googleAuds,
		AppleSignInClientIDs:         appleAuds,
	}
}

func splitComma(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseTTSLoudnessNorm 默认开启；设 XLANGAI_TTS_LOUDNESS_NORM=0|false|off 可关闭。
func parseTTSLoudnessNorm() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("XLANGAI_TTS_LOUDNESS_NORM")))
	switch v {
	case "0", "false", "no", "off":
		return false
	case "1", "true", "yes", "on":
		return true
	default:
		return true
	}
}

func parseTTSTargetLUFS() float64 {
	v := strings.TrimSpace(os.Getenv("XLANGAI_TTS_TARGET_LUFS"))
	if v == "" {
		return -16
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return -16
	}
	return f
}

func parseVerboseLogs() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("XLANGAI_VERBOSE_LOGS")))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return strings.ToLower(strings.TrimSpace(os.Getenv("GIN_MODE"))) != "release"
}

func loadLocalEnv() {
	paths := []string{
		".env",
		"env",
		filepath.Join("server", ".env"),
		filepath.Join("server", "env"),
	}

	for _, path := range paths {
		loadEnvFile(path)
	}
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		// 后加载的文件覆盖先加载的；不再因系统环境里已有同名变量就跳过，
		// 否则 Windows 上遗留的 DATABASE_URL 会让 .env 里的密码永远不生效。
		_ = os.Setenv(key, value)
	}
}

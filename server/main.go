package main

import (
	"context"

	"log"

	"net/url"

	"strings"

	"xlangai/server/config"

	"xlangai/server/internal/authz"

	"xlangai/server/internal/billing"

	"xlangai/server/internal/cache"

	"xlangai/server/internal/db"

	"xlangai/server/internal/handler"

	"xlangai/server/internal/loginotp"

	"xlangai/server/internal/media"

	appredis "xlangai/server/internal/redis"

	"xlangai/server/internal/repository"

	"xlangai/server/internal/router"

	"xlangai/server/internal/captcha"
	"xlangai/server/internal/settings"
	"xlangai/server/internal/sms"
)

func logDBTarget(dbURL string) {

	u, err := url.Parse(dbURL)

	if err != nil {

		log.Printf("DATABASE_URL 无法解析: %v", err)

		return

	}

	user := ""

	if u.User != nil {

		user = u.User.Username()

	}

	dbName := strings.TrimPrefix(u.Path, "/")

	host := u.Hostname()

	port := u.Port()

	if port == "" {

		port = "5432"

	}

	log.Printf("数据库连接: user=%q host=%s port=%s db=%q（密码来自 DATABASE_URL，错误 28P01 表示与 Postgres 里该用户密码不一致）", user, host, port, dbName)

}

func main() {

	cfg := config.Load()

	if cfg.VerboseLogs {
		log.Printf("XLANGAI 详细日志已启用：会话/STT 错误将打印路由与上下文；部分 JSON 响应含 detail（生产可设 GIN_MODE=release 或 XLANGAI_VERBOSE_LOGS=0）")
	}
	log.Printf("API 错误日志已启用：status>=400 → [api]（见 docs/server-logging.md）")

	logDBTarget(cfg.DBURL)

	ctx := context.Background()

	gdb, err := db.Open(ctx, cfg.DBURL)

	if err != nil {

		log.Fatal("db connect:", err)

	}

	sqlDB, err := gdb.DB()

	if err != nil {

		log.Fatal("db sql:", err)

	}

	defer sqlDB.Close()

	rdb := appredis.New(cfg.RedisURL)

	if rdb != nil {

		defer func() { _ = rdb.Close() }()

	}

	appCache := cache.Init(rdb)

	ur := repository.NewUserRepo(gdb)

	mr := repository.NewMembershipRepo(gdb)

	urUsage := repository.NewUsageRepo(gdb)

	az := authz.NewService(ur, mr, urUsage, appCache, cfg.PrincipalCacheTTL)

	br := repository.NewBillingRepo(gdb)

	var appleCfg *billing.AppleConfig

	if cfg.AppleIssuerID != "" && cfg.AppleKeyID != "" && cfg.ApplePrivateKeyPath != "" {

		key, err := billing.LoadApplePrivateKeyFromFile(cfg.ApplePrivateKeyPath)

		if err != nil {

			log.Printf("Apple IAP 私钥加载失败（将禁用 iOS 校验）: %v", err)

		} else {

			appleCfg = &billing.AppleConfig{

				IssuerID: cfg.AppleIssuerID,

				KeyID: cfg.AppleKeyID,

				BundleID: cfg.AppleBundleID,

				PrivateKey: key,

				Sandbox: cfg.AppleEnvironment != "production",
			}

		}

	}

	cr := repository.NewConvRepo(gdb)

	msgR := repository.NewMessageRepo(gdb)

	sr := repository.NewSystemRepo(gdb)

	llm := repository.NewLLMConfigRepo(gdb)

	stt := repository.NewSTTConfigRepo(gdb)

	tr := repository.NewTranslateConfigRepo(gdb)

	osr := repository.NewObjectStorageConfigRepo(gdb)

	smsR := repository.NewSmsConfigRepo(gdb)

	smsSvc := sms.NewService(smsR, cfg.VerboseLogs)

	ssr := repository.NewSystemSettingsRepo(gdb)

	sysSettings := settings.NewService(ssr)

	tc := repository.NewTtsConfigRepo(gdb)

	vr := repository.NewVoiceRepo(gdb)

	lr := repository.NewLangRepo(gdb)

	scenarioR := repository.NewScenarioRepo(gdb)

	openingR := repository.NewScenarioOpeningRepo(gdb)

	mediaSvc := media.NewService(osr, cfg, sysSettings)

	ah := handler.NewAIHandler(cfg, msgR, sr, cr, lr, llm, stt, tc, tr, vr, openingR, ur, urUsage, az, mediaSvc)

	ch := handler.NewConvHandler(cr, msgR, sr, scenarioR, ur, vr, lr, ah, cfg.VerboseLogs)

	scenarioH := handler.NewScenarioHandler(scenarioR)

	lh := handler.NewLangHandler(lr, appCache, cfg.LangCacheTTL)

	vh := handler.NewVoiceHandler(vr)

	otpStore := loginotp.NewStore(appCache)
	captchaStore := captcha.NewStore(appCache)

	captchaH := handler.NewCaptchaHandler(captchaStore)

	uh := handler.NewUserHandler(ur, lr, cfg, az, otpStore, captchaStore, smsSvc, mediaSvc, sysSettings)

	sth := handler.NewSettingsHandler(sysSettings)

	mdh := handler.NewMediaHandler(mediaSvc)

	mh := handler.NewMembershipHandler(mr)

	sh := handler.NewStatsHandler(repository.NewStatsRepo(gdb), urUsage)

	bh := handler.NewBillingHandler(cfg, appleCfg, br, ur, mr, az)

	r := router.New(cfg, az, uh, ch, ah, lh, vh, scenarioH, mh, sh, bh, sth, mdh, captchaH)

	log.Printf("listen :%s", cfg.Port)

	log.Fatal(r.Run(":" + cfg.Port))

}

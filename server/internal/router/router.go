package router

import (
	"wlltalk/server/config"
	"wlltalk/server/internal/authz"
	"wlltalk/server/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, az *authz.Service, uh *handler.UserHandler, ch *handler.ConvHandler, ah *handler.AIHandler, lh *handler.LangHandler, vh *handler.VoiceHandler, mh *handler.MembershipHandler, sh *handler.StatsHandler, bh *handler.BillingHandler, sth *handler.SettingsHandler, mdh *handler.MediaHandler) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery(), handler.HTTPErrorLogMiddleware(), gin.Logger())

	// Public
	r.GET("/api/v1/languages", lh.List)
	r.GET("/api/v1/public/settings", sth.PublicSettings)
	r.GET("/api/v1/membership/tiers", mh.ListTiers)
	r.GET("/api/v1/billing/catalog", bh.Catalog)
	r.POST("/api/v1/users", uh.Create)
	r.POST("/api/v1/auth/login", uh.Login)
	r.POST("/api/v1/auth/login/sms/send", uh.SendLoginSms)
	r.POST("/api/v1/auth/login/sms", uh.LoginWithSms)
	r.POST("/api/v1/auth/register/sms/send", uh.SendRegisterSms)
	r.POST("/api/v1/auth/register/sms", uh.RegisterWithSms)
	r.POST("/api/v1/auth/login/google", uh.LoginGoogle)
	r.POST("/api/v1/auth/login/apple", uh.LoginApple)

	// 公开：音频文件（TTS 生成）
	r.GET("/api/v1/audio/:filename", ah.ServeAudio)
	// 公开：用户头像（路径为随机文件名）
	r.GET("/api/v1/avatars/:filename", uh.ServeAvatar)

	// Protected
	prot := r.Group("/api/v1")
	prot.Use(handler.AuthMiddleware(cfg), handler.PrincipalMiddleware(az))
	{
		prot.GET("/users/me", uh.Me)
		prot.PATCH("/users/me", uh.PatchMe)
		prot.POST("/users/me/avatar", uh.UploadAvatar)
		prot.GET("/media/capabilities", mdh.Capabilities)
		prot.POST("/media/upload-presign", mdh.PresignUpload)
		prot.GET("/stats/summary", sh.Summary)
		prot.GET("/stats/calendar", sh.Calendar)
		prot.POST("/billing/verify", bh.Verify)
		prot.GET("/voices", vh.List)
		prot.POST("/conversations", ch.Create)
		prot.GET("/conversations", ch.List)
		prot.GET("/conversations/:id", ch.Get)
		prot.PATCH("/conversations/:id", ch.Patch)
		prot.DELETE("/conversations/:id", ch.Delete)
		prot.PUT("/conversations/:id/voice", ch.UpdateVoice)
		prot.GET("/conversations/:id/messages", ch.ListMessages)
		prot.POST("/conversations/:id/chat", ah.Chat)
		prot.POST("/conversations/:id/voice", ah.VoiceChat)
		prot.POST("/translate", ah.Translate)
	}

	return r
}

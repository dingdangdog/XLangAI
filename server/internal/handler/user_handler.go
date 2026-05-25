package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"xlangai/server/config"
	"xlangai/server/internal/auth"
	"xlangai/server/internal/authz"
	"xlangai/server/internal/loginotp"
	"xlangai/server/internal/media"
	"xlangai/server/internal/model"
	"xlangai/server/internal/oauth"
	"xlangai/server/internal/objectstore"
	"xlangai/server/internal/repository"
	"xlangai/server/internal/settings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	repo     *repository.UserRepo
	langRepo *repository.LangRepo
	cfg      *config.Config
	az       *authz.Service
	otp      *loginotp.Store
	media    *media.Service
	settings *settings.Service
}

func NewUserHandler(repo *repository.UserRepo, langRepo *repository.LangRepo, cfg *config.Config, az *authz.Service, otp *loginotp.Store, mediaSvc *media.Service, sys *settings.Service) *UserHandler {
	return &UserHandler{repo: repo, langRepo: langRepo, cfg: cfg, az: az, otp: otp, media: mediaSvc, settings: sys}
}

func authAPIError(c *gin.Context, status int, msg, code string) {
	c.JSON(status, gin.H{"error": msg, "code": code})
}

func (h *UserHandler) authEnabled(c *gin.Context, key string) bool {
	if h.settings == nil || !h.settings.Bool(c.Request.Context(), key) {
		authAPIError(c, http.StatusForbidden, "login method disabled", "LOGIN_METHOD_DISABLED")
		return false
	}
	return true
}

func (h *UserHandler) Create(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthPasswordEnabled) {
		return
	}
	if !h.authEnabled(c, settings.AuthPasswordRegisterEnabled) {
		return
	}
	var req struct {
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Phone == "" && req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone or email required"})
		return
	}
	u, err := h.repo.Create(c.Request.Context(), repository.CreateUserParams{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toUserResp(c.Request.Context(), h.langRepo, u))
}

func (h *UserHandler) Login(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthPasswordEnabled) {
		return
	}
	var req struct {
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.repo.GetByPhoneOrEmail(c.Request.Context(), req.Phone, req.Email)
	if err != nil {
		authAPIError(c, http.StatusUnauthorized, "invalid credentials", "INVALID_CREDENTIALS")
		return
	}
	if !h.repo.CheckPassword(u.PasswordHash, req.Password) {
		authAPIError(c, http.StatusUnauthorized, "invalid credentials", "INVALID_CREDENTIALS")
		return
	}
	role := u.Role
	if role == "" {
		role = authz.RoleUser
	}
	token, err := auth.GenerateToken(u.ID, role, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": toUserResp(c.Request.Context(), h.langRepo, u)})
}

const loginOtpTTL = 5 * time.Minute
const loginSmsCooldown = 55 * time.Second

// SendLoginSms 发送登录验证码（当前未对接短信网关；验证码存 Redis/内存，开启详细日志时打印到控制台便于联调）。
func (h *UserHandler) SendLoginSms(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthSmsEnabled) {
		return
	}
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if len(phone) < 8 {
		authAPIError(c, http.StatusBadRequest, "invalid phone", "INVALID_PHONE")
		return
	}
	ctx := c.Request.Context()
	if h.otp.CooldownActive(ctx, phone) {
		authAPIError(c, http.StatusTooManyRequests, "please wait before resending", "SMS_COOLDOWN")
		return
	}
	_, err := h.repo.GetByPhone(ctx, phone)
	if err != nil {
		authAPIError(c, http.StatusNotFound, "phone not registered", "PHONE_NOT_REGISTERED")
		return
	}
	code := loginotp.RandomDigits6()
	h.otp.PutCode(ctx, phone, code, loginOtpTTL)
	h.otp.SetCooldown(ctx, phone, loginSmsCooldown)
	if h.cfg.VerboseLogs {
		log.Printf("[auth] login OTP phone=%s code=%s (SMS not integrated; shown for debugging only)\n", phone, code)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// LoginWithSms 使用手机号 + 短信验证码登录。
func (h *UserHandler) LoginWithSms(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthSmsEnabled) {
		return
	}
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	code := strings.TrimSpace(req.Code)
	ctx := c.Request.Context()
	want, ok := h.otp.GetCode(ctx, phone)
	if !ok || want != code {
		authAPIError(c, http.StatusUnauthorized, "invalid or expired code", "INVALID_OR_EXPIRED_CODE")
		return
	}
	u, err := h.repo.GetByPhone(ctx, phone)
	if err != nil {
		authAPIError(c, http.StatusUnauthorized, "invalid credentials", "INVALID_CREDENTIALS")
		return
	}
	h.otp.DeleteCode(ctx, phone)
	role := u.Role
	if role == "" {
		role = authz.RoleUser
	}
	token, err := auth.GenerateToken(u.ID, role, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": toUserResp(c.Request.Context(), h.langRepo, u)})
}

// SendRegisterSms 向未注册手机号发送注册验证码。
func (h *UserHandler) SendRegisterSms(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthSmsRegisterEnabled) {
		return
	}
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if len(phone) < 8 {
		authAPIError(c, http.StatusBadRequest, "invalid phone", "INVALID_PHONE")
		return
	}
	ctx := c.Request.Context()
	if h.otp.RegisterCooldownActive(ctx, phone) {
		authAPIError(c, http.StatusTooManyRequests, "please wait before resending", "SMS_COOLDOWN")
		return
	}
	if _, err := h.repo.GetByPhone(ctx, phone); err == nil {
		authAPIError(c, http.StatusConflict, "phone already registered", "PHONE_ALREADY_REGISTERED")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	code := loginotp.RandomDigits6()
	h.otp.PutRegisterCode(ctx, phone, code, loginOtpTTL)
	h.otp.SetRegisterCooldown(ctx, phone, loginSmsCooldown)
	if h.cfg.VerboseLogs {
		log.Printf("[auth] register OTP phone=%s code=%s (SMS not integrated; shown for debugging only)\n", phone, code)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// RegisterWithSms 使用手机号 + 验证码完成注册并登录。
func (h *UserHandler) RegisterWithSms(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthSmsRegisterEnabled) {
		return
	}
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	code := strings.TrimSpace(req.Code)
	ctx := c.Request.Context()
	want, ok := h.otp.GetRegisterCode(ctx, phone)
	if !ok || want != code {
		authAPIError(c, http.StatusUnauthorized, "invalid or expired code", "INVALID_OR_EXPIRED_CODE")
		return
	}
	if _, err := h.repo.GetByPhone(ctx, phone); err == nil {
		authAPIError(c, http.StatusConflict, "phone already registered", "PHONE_ALREADY_REGISTERED")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	u, err := h.repo.CreateSmsUser(ctx, phone, req.Nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.otp.DeleteRegisterCode(ctx, phone)
	h.oauthSessionJSON(c, u)
}

func (h *UserHandler) oauthSessionJSON(c *gin.Context, u *model.User) {
	role := u.Role
	if role == "" {
		role = authz.RoleUser
	}
	token, err := auth.GenerateToken(u.ID, role, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": toUserResp(c.Request.Context(), h.langRepo, u)})
}

// LoginGoogle 使用 Google id_token 登录或注册（首次自动建号）。
func (h *UserHandler) LoginGoogle(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthGoogleEnabled) {
		return
	}
	if len(h.cfg.GoogleOAuthClientIDs) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "google login not configured"})
		return
	}
	var req struct {
		IDToken string `json:"id_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	sub, emailPtr, err := oauth.VerifyGoogleIDToken(ctx, req.IDToken, h.cfg.GoogleOAuthClientIDs)
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[auth] google id_token invalid: %v\n", err)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid google token"})
		return
	}
	u, err := h.repo.GetByGoogleSub(ctx, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if !h.settings.Bool(ctx, settings.AuthGoogleRegisterEnabled) {
				c.JSON(http.StatusForbidden, gin.H{"error": "registration disabled"})
				return
			}
			u, err = h.repo.CreateOAuthUser(ctx, repository.CreateOAuthUserParams{
				GoogleSub: &sub,
				Email:     emailPtr,
				Nickname:  nil,
			})
		}
	}
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[auth] google login db: %v\n", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	h.oauthSessionJSON(c, u)
}

// LoginApple 使用 Apple identityToken（JWT）登录或注册。
func (h *UserHandler) LoginApple(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthAppleEnabled) {
		return
	}
	if len(h.cfg.AppleSignInClientIDs) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "apple login not configured"})
		return
	}
	var req struct {
		IdentityToken string `json:"identity_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	sub, emailPtr, err := oauth.VerifyAppleIDToken(ctx, req.IdentityToken, h.cfg.AppleSignInClientIDs)
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[auth] apple identity_token invalid: %v\n", err)
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid apple token"})
		return
	}
	u, err := h.repo.GetByAppleSub(ctx, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if !h.settings.Bool(ctx, settings.AuthAppleRegisterEnabled) {
				c.JSON(http.StatusForbidden, gin.H{"error": "registration disabled"})
				return
			}
			u, err = h.repo.CreateOAuthUser(ctx, repository.CreateOAuthUserParams{
				AppleSub: &sub,
				Email:    emailPtr,
				Nickname: nil,
			})
		}
	}
	if err != nil {
		if h.cfg.VerboseLogs {
			log.Printf("[auth] apple login db: %v\n", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	h.oauthSessionJSON(c, u)
}

func (h *UserHandler) Me(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	p := CtxPrincipal(c)
	u, err := h.repo.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	resp := toUserResp(c.Request.Context(), h.langRepo, u)
	if p != nil {
		if p.Role != "" {
			resp["role"] = p.Role
		}
		if p.Tier != nil {
			resp["tier"] = gin.H{
				"id":            p.Tier.ID,
				"code":          p.Tier.Code,
				"name":          p.Tier.Name,
				"daily_limit":   p.Tier.DailyLimit,
				"monthly_limit": p.Tier.MonthlyLimit,
			}
		}
	}
	if h.az != nil {
		today, month, err := h.az.MeUsage(c.Request.Context(), uid)
		if err == nil {
			resp["usage"] = gin.H{
				"chat_today":      today,
				"chat_this_month": month,
			}
		}
	}
	resp["token_balance"] = u.TokenBalance
	if u.SubscriptionExpiresAt != nil {
		resp["subscription_expires_at"] = u.SubscriptionExpiresAt.UTC().Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, resp)
}

const maxAvatarUploadBytes = 2 << 20

// PatchMe 更新当前用户资料：nickname、native_language_id（可单独或同时提交；传 JSON null 清除母语）。
func (h *UserHandler) PatchMe(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, hasNick := raw["nickname"]
	_, hasNative := raw["native_language_id"]
	if !hasNick && !hasNative {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nickname or native_language_id required"})
		return
	}

	var nickPtr, nativePtr *string
	if hasNick {
		rawNick := raw["nickname"]
		var nick string
		if len(rawNick) > 0 && string(rawNick) != "null" {
			if err := json.Unmarshal(rawNick, &nick); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid nickname"})
				return
			}
		}
		if utf8.RuneCountInString(strings.TrimSpace(nick)) > 32 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nickname too long (max 32)"})
			return
		}
		nickPtr = &nick
	}
	if hasNative {
		rawNative := raw["native_language_id"]
		if len(rawNative) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid native_language_id"})
			return
		}
		if string(rawNative) == "null" {
			empty := ""
			nativePtr = &empty
		} else {
			var nativeID string
			if err := json.Unmarshal(rawNative, &nativeID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid native_language_id"})
				return
			}
			nativeID = strings.TrimSpace(nativeID)
			if nativeID != "" {
				if h.langRepo == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "language service unavailable"})
					return
				}
				if _, err := h.langRepo.GetCodeByID(c.Request.Context(), nativeID); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "unknown native_language_id"})
					return
				}
			}
			nativePtr = &nativeID
		}
	}

	if err := h.repo.UpdateProfile(c.Request.Context(), uid, nickPtr, nil, nativePtr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.Me(c)
}

// ChangePassword 修改当前用户密码（需已设置密码；OAuth/短信注册用户无密码时不可用）。
func (h *UserHandler) ChangePassword(c *gin.Context) {
	if !h.authEnabled(c, settings.AuthPasswordEnabled) {
		return
	}
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if u.PasswordHash == "" {
		authAPIError(c, http.StatusBadRequest, "password not set", "PASSWORD_NOT_SET")
		return
	}
	if !h.repo.CheckPassword(u.PasswordHash, req.OldPassword) {
		authAPIError(c, http.StatusUnauthorized, "invalid current password", "INVALID_CURRENT_PASSWORD")
		return
	}
	if req.OldPassword == req.NewPassword {
		authAPIError(c, http.StatusBadRequest, "new password must differ from current password", "PASSWORD_SAME_AS_OLD")
		return
	}
	if err := h.repo.UpdatePassword(c.Request.Context(), uid, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UploadAvatar 接收 multipart 字段 `file` 上传，或 JSON/form 字段 `avatar_url`（客户端已直传 R2 后确认）。
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	confirmURL := strings.TrimSpace(c.PostForm("avatar_url"))
	if confirmURL == "" {
		var body struct {
			AvatarURL string `json:"avatar_url"`
		}
		if c.ShouldBindJSON(&body) == nil {
			confirmURL = strings.TrimSpace(body.AvatarURL)
		}
	}
	if confirmURL != "" {
		if err := h.media.ValidateObjectURL(c.Request.Context(), media.ScopeAvatar, confirmURL); err != nil {
			if errors.Is(err, media.ErrInvalidObjectURL) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar_url"})
				return
			}
			writeMediaStorageError(c, err)
			return
		}
		if err := h.repo.UpdateProfile(c.Request.Context(), uid, nil, &confirmURL, nil); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"avatar_url": confirmURL})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read file"})
		return
	}
	defer src.Close()
	data, err := io.ReadAll(io.LimitReader(src, maxAvatarUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read failed"})
		return
	}
	if len(data) > maxAvatarUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (max 2MB)"})
		return
	}
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(file.Filename)))
	if !allowedAvatarExt(ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type"})
		return
	}
	res, err := h.media.SaveAvatar(c.Request.Context(), data, ext, avatarMimeType("x"+ext))
	if err != nil {
		if errors.Is(err, media.ErrClientOnlyStorage) {
			c.JSON(http.StatusForbidden, gin.H{"error": "avatar storage is client-only"})
			return
		}
		if errors.Is(err, objectstore.ErrNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "对象存储未配置或凭证不完整，请在管理后台配置并启用"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	url := res.URL
	if err := h.repo.UpdateProfile(c.Request.Context(), uid, nil, &url, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar_url": url})
}

// ServeAvatar 公开读取头像文件（文件名含随机 UUID，与 TTS 音频类似）。
func (h *UserHandler) ServeAvatar(c *gin.Context) {
	name := c.Param("filename")
	if name == "" || len(name) > 120 || strings.ContainsAny(name, "/\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}
	data, err := h.media.ReadAvatar(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Data(http.StatusOK, avatarMimeType(name), data)
}

func allowedAvatarExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	default:
		return false
	}
}

func avatarMimeType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func toUserResp(ctx context.Context, langRepo *repository.LangRepo, u *model.User) gin.H {
	resp := gin.H{"id": u.ID, "status": u.Status, "has_password": u.PasswordHash != ""}
	if u.Role != "" {
		resp["role"] = u.Role
	}
	if u.Phone != nil {
		resp["phone"] = *u.Phone
	}
	if u.Email != nil {
		resp["email"] = *u.Email
	}
	if u.Nickname != nil {
		resp["nickname"] = *u.Nickname
	}
	if u.AvatarURL != nil && strings.TrimSpace(*u.AvatarURL) != "" {
		resp["avatar_url"] = *u.AvatarURL
	}
	if u.LanguageID != nil && strings.TrimSpace(*u.LanguageID) != "" {
		id := strings.TrimSpace(*u.LanguageID)
		resp["native_language_id"] = id
		if langRepo != nil {
			if code, err := langRepo.GetCodeByID(ctx, id); err == nil {
				resp["native_language_code"] = code
			}
		}
	}
	return resp
}

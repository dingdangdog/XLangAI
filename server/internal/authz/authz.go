package authz

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"xlangai/server/internal/cache"
	"xlangai/server/internal/model"
	"xlangai/server/internal/repository"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var (
	ErrForbiddenLanguage = errors.New("membership does not allow this language")
	ErrForbiddenVoice    = errors.New("membership does not allow this voice role")
	ErrQuotaDaily        = errors.New("daily chat limit reached")
	ErrQuotaMonthly      = errors.New("monthly chat limit reached")
	ErrQuotaTokens       = errors.New("monthly included turns exhausted and token balance is empty")
	ErrQuotaTurns        = errors.New("permanent turn balance is empty")
)

type tierFeatures struct {
	LanguageIDs  []string `json:"language_ids"`
	VoiceRoleIDs []string `json:"voice_role_ids"`
}

// Principal 当前请求用户及其会员能力（由中间件注入 Gin）。
type Principal struct {
	UserID       string                     `json:"user_id"`
	Role         string                     `json:"role"`
	Tier         *repository.MembershipTier `json:"tier,omitempty"`
	TokenBalance int64                      `json:"token_balance"`
	TurnBalance  int                        `json:"turn_balance"`
	Feat         tierFeatures               `json:"-"`
}

type cachedPrincipal struct {
	UserID       string                     `json:"user_id"`
	Role         string                     `json:"role"`
	Tier         *repository.MembershipTier `json:"tier,omitempty"`
	TokenBalance int64                      `json:"token_balance"`
	TurnBalance  int                        `json:"turn_balance"`
	LanguageIDs  []string                   `json:"language_ids,omitempty"`
	VoiceRoleIDs []string                   `json:"voice_role_ids,omitempty"`
}

type Service struct {
	users *repository.UserRepo
	tiers *repository.MembershipRepo
	usage *repository.UsageRepo
	cache *cache.Cache
	ttl   cacheTTL
}

type cacheTTL struct {
	Principal time.Duration
}

func NewService(users *repository.UserRepo, tiers *repository.MembershipRepo, usage *repository.UsageRepo, c *cache.Cache, principalTTL time.Duration) *Service {
	if principalTTL <= 0 {
		principalTTL = 60 * time.Second
	}
	return &Service{
		users: users,
		tiers: tiers,
		usage: usage,
		cache: c,
		ttl:   cacheTTL{Principal: principalTTL},
	}
}

func parseFeatures(raw *string) tierFeatures {
	var f tierFeatures
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return f
	}
	_ = json.Unmarshal([]byte(*raw), &f)
	return f
}

func (s *Service) principalFromUserTier(u *model.User, t *repository.MembershipTier) *Principal {
	role := u.Role
	if role == "" {
		role = RoleUser
	}
	p := &Principal{
		UserID:       u.ID,
		Role:         role,
		Tier:         t,
		TokenBalance: u.TokenBalance,
		TurnBalance:  u.TurnBalance,
	}
	if t != nil {
		p.Feat = parseFeatures(t.FeaturesJSON)
	}
	return p
}

func (s *Service) toCached(p *Principal) cachedPrincipal {
	return cachedPrincipal{
		UserID:       p.UserID,
		Role:         p.Role,
		Tier:         p.Tier,
		TokenBalance: p.TokenBalance,
		TurnBalance:  p.TurnBalance,
		LanguageIDs:  p.Feat.LanguageIDs,
		VoiceRoleIDs: p.Feat.VoiceRoleIDs,
	}
}

func (s *Service) fromCached(c *cachedPrincipal) *Principal {
	if c == nil {
		return nil
	}
	p := &Principal{
		UserID:       c.UserID,
		Role:         c.Role,
		Tier:         c.Tier,
		TokenBalance: c.TokenBalance,
		TurnBalance:  c.TurnBalance,
	}
	p.Feat.LanguageIDs = c.LanguageIDs
	p.Feat.VoiceRoleIDs = c.VoiceRoleIDs
	if p.Role == "" {
		p.Role = RoleUser
	}
	return p
}

// LoadPrincipal 加载用户与会员档位（带 Redis 缓存）。
func (s *Service) LoadPrincipal(ctx context.Context, userID string) (*Principal, error) {
	key := cache.PrincipalKey(userID)
	var cp cachedPrincipal
	if s.cache != nil && s.cache.GetJSON(ctx, key, &cp) {
		return s.fromCached(&cp), nil
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var tier *repository.MembershipTier
	if u.TierID != nil && *u.TierID != "" {
		tier, err = s.tiers.GetByID(ctx, *u.TierID)
		if err != nil {
			return nil, err
		}
	}
	if tier == nil {
		fid, _ := s.tiers.GetFreeTierID(ctx)
		if fid != nil {
			tier, _ = s.tiers.GetByID(ctx, *fid)
		}
	}
	// 商店订阅已过期：内存中按免费档额度控制（不立刻写回 tier_id，避免高频 UPDATE）。
	if tier != nil && u.SubscriptionExpiresAt != nil {
		if time.Now().UTC().After(*u.SubscriptionExpiresAt) {
			code := strings.TrimSpace(strings.ToLower(tier.Code))
			if code == "plus" || code == "pro" {
				fid, _ := s.tiers.GetFreeTierID(ctx)
				if fid != nil {
					tier, _ = s.tiers.GetByID(ctx, *fid)
				}
			}
		}
	}
	p := s.principalFromUserTier(u, tier)
	if s.cache != nil {
		s.cache.SetJSON(ctx, key, s.toCached(p), s.ttl.Principal)
	}
	return p, nil
}

func (s *Service) InvalidatePrincipal(ctx context.Context, userID string) {
	if s.cache == nil {
		return
	}
	s.cache.Delete(ctx, cache.PrincipalKey(userID))
}

// EnsureLanguageAllowed 校验会话/练习目标语言是否在会员允许列表内（features 为空则不限）。
func (p *Principal) EnsureLanguageAllowed(languageID string) error {
	if p == nil || languageID == "" {
		return nil
	}
	ids := p.Feat.LanguageIDs
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if id == languageID {
			return nil
		}
	}
	return ErrForbiddenLanguage
}

// EnsureVoiceAllowed 校验音色是否在会员允许列表内（features 为空或 voiceID 为空则跳过）。
func (p *Principal) EnsureVoiceAllowed(voiceRoleID *string) error {
	if p == nil || voiceRoleID == nil || *voiceRoleID == "" {
		return nil
	}
	ids := p.Feat.VoiceRoleIDs
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if id == *voiceRoleID {
			return nil
		}
	}
	return ErrForbiddenVoice
}

// QuotaErrorCode 将额度错误映射为 API code（空串表示非额度错误）。
func QuotaErrorCode(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ErrQuotaDaily):
		return "QUOTA_DAILY"
	case errors.Is(err, ErrQuotaMonthly):
		return "QUOTA_MONTHLY"
	case errors.Is(err, ErrQuotaTokens):
		return "QUOTA_TOKENS"
	case errors.Is(err, ErrQuotaTurns):
		return "QUOTA_TURNS"
	default:
		return ""
	}
}

// UsesTurnBalance 为 true 时：有效档位无日/月日历限额，改用永久次数余额。
func (p *Principal) UsesTurnBalance() bool {
	if p == nil {
		return true
	}
	if p.Tier == nil {
		return true
	}
	return p.Tier.DailyLimit <= 0 && p.Tier.MonthlyLimit <= 0
}

// EnsureChatQuota 发送对话前校验额度。
// - 无日/月限额的档位（如 free）：检查 turn_balance
// - 有日/月限额的档位（如 plus/pro）：沿用日历限额；月满后可用 token_balance 兜底
func (s *Service) EnsureChatQuota(ctx context.Context, p *Principal) error {
	if p == nil {
		return nil
	}
	if p.UsesTurnBalance() {
		if p.TurnBalance > 0 {
			return nil
		}
		return ErrQuotaTurns
	}
	daily := p.Tier.DailyLimit
	monthly := p.Tier.MonthlyLimit
	if daily > 0 {
		n, err := s.usage.TodayUsageCount(ctx, p.UserID)
		if err != nil {
			return err
		}
		if n >= daily {
			return ErrQuotaDaily
		}
	}
	if monthly > 0 {
		n, err := s.usage.MonthUsageCount(ctx, p.UserID)
		if err != nil {
			return err
		}
		if n < monthly {
			return nil
		}
		if p.TokenBalance > 0 {
			return nil
		}
		return ErrQuotaTokens
	}
	return nil
}

// UsesTokenWalletForNextTurn 为 true 时，本回合完成后应按 LLM token 从钱包扣费（在仍通过 EnsureChatQuota 的前提下）。
func (s *Service) UsesTokenWalletForNextTurn(ctx context.Context, p *Principal) (bool, error) {
	if p == nil || p.UsesTurnBalance() || p.Tier == nil {
		return false, nil
	}
	if p.Tier.MonthlyLimit <= 0 {
		return false, nil
	}
	n, err := s.usage.MonthUsageCount(ctx, p.UserID)
	if err != nil {
		return false, err
	}
	return n >= p.Tier.MonthlyLimit, nil
}

// UsesTurnWalletForNextTurn 为 true 时，本回合完成后应从永久次数余额扣 1。
func (s *Service) UsesTurnWalletForNextTurn(p *Principal) bool {
	return p != nil && p.UsesTurnBalance()
}

// DeductChatTokens 对话完成后按 LLM 用量扣减 token 余额（仅影响 usr_users.token_balance）。
func (s *Service) DeductChatTokens(ctx context.Context, userID string, n int64) error {
	if n <= 0 {
		return nil
	}
	_, err := s.users.DeductTokenBalance(ctx, userID, n)
	if err == nil {
		s.InvalidatePrincipal(ctx, userID)
	}
	return err
}

// DeductChatTurn 对话完成后扣减永久次数余额 1。
func (s *Service) DeductChatTurn(ctx context.Context, userID string) error {
	ok, err := s.users.DeductTurnBalance(ctx, userID, 1)
	if err != nil {
		return err
	}
	if ok {
		s.InvalidatePrincipal(ctx, userID)
	}
	return nil
}

func (s *Service) RecordChatTurn(ctx context.Context, userID string) error {
	return s.usage.AddChatTurn(ctx, userID)
}

// MeUsage 今日与本月已用轮次（用于 /users/me）。
func (s *Service) MeUsage(ctx context.Context, userID string) (today, month int, err error) {
	today, err = s.usage.TodayUsageCount(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	month, err = s.usage.MonthUsageCount(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return today, month, nil
}

func (p *Principal) IsAdmin() bool {
	return p != nil && p.Role == RoleAdmin
}

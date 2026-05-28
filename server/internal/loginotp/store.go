package loginotp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"xlangai/server/internal/cache"
)

const (
	otpKeyFmt         = "xlangai:login_otp:v1:%s"
	cdKeyFmt          = "xlangai:login_otp_cd:v1:%s"
	registerOtpKeyFmt = "xlangai:register_otp:v1:%s"
	registerCdKeyFmt  = "xlangai:register_otp_cd:v1:%s"
)

// Store 登录短信验证码；底层走统一 cache（Redis 或进程内内存）。
type Store struct {
	c *cache.Cache
}

func NewStore(c *cache.Cache) *Store {
	return &Store{c: c}
}

func normPhone(p string) string { return strings.TrimSpace(p) }

func otpKey(phone string) string { return fmt.Sprintf(otpKeyFmt, phone) }
func cdKey(phone string) string  { return fmt.Sprintf(cdKeyFmt, phone) }

func registerOtpKey(phone string) string { return fmt.Sprintf(registerOtpKeyFmt, phone) }
func registerCdKey(phone string) string  { return fmt.Sprintf(registerCdKeyFmt, phone) }

func (s *Store) CooldownActive(ctx context.Context, phone string) bool {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return false
	}
	_, ok := s.c.GetPlain(ctx, cdKey(p))
	return ok
}

func (s *Store) SetCooldown(ctx context.Context, phone string, d time.Duration) {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return
	}
	_ = s.c.SetPlain(ctx, cdKey(p), "1", d)
}

func (s *Store) PutCode(ctx context.Context, phone, code string, ttl time.Duration) {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return
	}
	_ = s.c.SetPlain(ctx, otpKey(p), code, ttl)
}

func (s *Store) GetCode(ctx context.Context, phone string) (string, bool) {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return "", false
	}
	return s.c.GetPlain(ctx, otpKey(p))
}

func (s *Store) DeleteCode(ctx context.Context, phone string) {
	p := normPhone(phone)
	if s.c == nil {
		return
	}
	s.c.Delete(ctx, otpKey(p))
}

func (s *Store) RegisterCooldownActive(ctx context.Context, phone string) bool {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return false
	}
	_, ok := s.c.GetPlain(ctx, registerCdKey(p))
	return ok
}

func (s *Store) SetRegisterCooldown(ctx context.Context, phone string, d time.Duration) {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return
	}
	_ = s.c.SetPlain(ctx, registerCdKey(p), "1", d)
}

func (s *Store) PutRegisterCode(ctx context.Context, phone, code string, ttl time.Duration) {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return
	}
	_ = s.c.SetPlain(ctx, registerOtpKey(p), code, ttl)
}

func (s *Store) GetRegisterCode(ctx context.Context, phone string) (string, bool) {
	p := normPhone(phone)
	if p == "" || s.c == nil {
		return "", false
	}
	return s.c.GetPlain(ctx, registerOtpKey(p))
}

func (s *Store) DeleteRegisterCode(ctx context.Context, phone string) {
	p := normPhone(phone)
	if s.c == nil {
		return
	}
	s.c.Delete(ctx, registerOtpKey(p))
}

// RandomDigits6 生成 6 位数字验证码。
func RandomDigits6() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

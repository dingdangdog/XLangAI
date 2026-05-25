package loginotp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"xlangai/server/internal/cache"
)

const (
	otpKeyFmt         = "xlangai:login_otp:v1:%s"
	cdKeyFmt          = "xlangai:login_otp_cd:v1:%s"
	registerOtpKeyFmt = "xlangai:register_otp:v1:%s"
	registerCdKeyFmt  = "xlangai:register_otp_cd:v1:%s"
)

// Store 登录短信验证码：优先 Redis（与 cache.Cache 共用客户端），否则进程内内存（单实例开发可用）。
type Store struct {
	c   *cache.Cache
	mem *memStore
}

type memStore struct {
	mu  sync.Mutex
	otp map[string]memVal
	cd  map[string]time.Time
}

type memVal struct {
	code  string
	until time.Time
}

func NewStore(c *cache.Cache) *Store {
	return &Store{c: c, mem: &memStore{otp: make(map[string]memVal), cd: make(map[string]time.Time)}}
}

func normPhone(p string) string { return strings.TrimSpace(p) }

func otpKey(phone string) string { return fmt.Sprintf(otpKeyFmt, phone) }
func cdKey(phone string) string  { return fmt.Sprintf(cdKeyFmt, phone) }

func registerOtpKey(phone string) string { return fmt.Sprintf(registerOtpKeyFmt, phone) }
func registerCdKey(phone string) string  { return fmt.Sprintf(registerCdKeyFmt, phone) }

func (s *Store) CooldownActive(ctx context.Context, phone string) bool {
	p := normPhone(phone)
	if p == "" {
		return false
	}
	if s.c != nil && s.c.HasRedis() {
		if _, ok := s.c.GetPlain(ctx, cdKey(p)); ok {
			return true
		}
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	until, ok := s.mem.cd[p]
	return ok && time.Now().Before(until)
}

func (s *Store) SetCooldown(ctx context.Context, phone string, d time.Duration) {
	p := normPhone(phone)
	if p == "" {
		return
	}
	if s.c != nil && s.c.HasRedis() && s.c.SetPlain(ctx, cdKey(p), "1", d) {
		return
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	s.mem.cd[p] = time.Now().Add(d)
}

func (s *Store) PutCode(ctx context.Context, phone, code string, ttl time.Duration) {
	p := normPhone(phone)
	if p == "" {
		return
	}
	if s.c != nil && s.c.HasRedis() && s.c.SetPlain(ctx, otpKey(p), code, ttl) {
		return
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	s.mem.otp[p] = memVal{code: code, until: time.Now().Add(ttl)}
}

func (s *Store) GetCode(ctx context.Context, phone string) (string, bool) {
	p := normPhone(phone)
	if p == "" {
		return "", false
	}
	if s.c != nil && s.c.HasRedis() {
		return s.c.GetPlain(ctx, otpKey(p))
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	v, ok := s.mem.otp[p]
	if !ok || time.Now().After(v.until) {
		delete(s.mem.otp, p)
		return "", false
	}
	return v.code, true
}

func (s *Store) DeleteCode(ctx context.Context, phone string) {
	p := normPhone(phone)
	if s.c != nil && s.c.HasRedis() {
		s.c.Delete(ctx, otpKey(p))
		return
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	delete(s.mem.otp, p)
}

func (s *Store) RegisterCooldownActive(ctx context.Context, phone string) bool {
	p := normPhone(phone)
	if p == "" {
		return false
	}
	if s.c != nil && s.c.HasRedis() {
		if _, ok := s.c.GetPlain(ctx, registerCdKey(p)); ok {
			return true
		}
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	until, ok := s.mem.cd["reg:"+p]
	return ok && time.Now().Before(until)
}

func (s *Store) SetRegisterCooldown(ctx context.Context, phone string, d time.Duration) {
	p := normPhone(phone)
	if p == "" {
		return
	}
	if s.c != nil && s.c.HasRedis() && s.c.SetPlain(ctx, registerCdKey(p), "1", d) {
		return
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	s.mem.cd["reg:"+p] = time.Now().Add(d)
}

func (s *Store) PutRegisterCode(ctx context.Context, phone, code string, ttl time.Duration) {
	p := normPhone(phone)
	if p == "" {
		return
	}
	if s.c != nil && s.c.HasRedis() && s.c.SetPlain(ctx, registerOtpKey(p), code, ttl) {
		return
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	s.mem.otp["reg:"+p] = memVal{code: code, until: time.Now().Add(ttl)}
}

func (s *Store) GetRegisterCode(ctx context.Context, phone string) (string, bool) {
	p := normPhone(phone)
	if p == "" {
		return "", false
	}
	if s.c != nil && s.c.HasRedis() {
		return s.c.GetPlain(ctx, registerOtpKey(p))
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	v, ok := s.mem.otp["reg:"+p]
	if !ok || time.Now().After(v.until) {
		delete(s.mem.otp, "reg:"+p)
		return "", false
	}
	return v.code, true
}

func (s *Store) DeleteRegisterCode(ctx context.Context, phone string) {
	p := normPhone(phone)
	if s.c != nil && s.c.HasRedis() {
		s.c.Delete(ctx, registerOtpKey(p))
		return
	}
	s.mem.mu.Lock()
	defer s.mem.mu.Unlock()
	delete(s.mem.otp, "reg:"+p)
}

// RandomDigits6 生成 6 位数字验证码。
func RandomDigits6() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

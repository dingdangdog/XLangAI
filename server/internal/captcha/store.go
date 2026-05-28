package captcha

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"xlangai/server/internal/cache"
)

const (
	keyFmt           = "xlangai:captcha:v1:%s"
	ticketTTL        = 120 * time.Second
	verifiedConsume  = 60 * time.Second
)

var (
	ErrTicketNotFound   = errors.New("captcha ticket not found")
	ErrNotVerified      = errors.New("captcha not verified")
	ErrAlreadyConsumed  = errors.New("captcha already consumed")
	ErrWrongAnswer      = errors.New("wrong captcha answer")
	ErrInvalidAnswer    = errors.New("invalid captcha answer")
)

type record struct {
	Question string  `json:"question"`
	Answer   float64 `json:"answer"`
	Verified bool    `json:"verified"`
}

// Store 数学题人机验证票据（Redis / 内存）。
type Store struct {
	c *cache.Cache
}

func NewStore(c *cache.Cache) *Store {
	return &Store{c: c}
}

func key(ticket string) string {
	return fmt.Sprintf(keyFmt, ticket)
}

// CreateTicket 创建新题目，返回 ticket、题目文案、有效秒数。
func (s *Store) CreateTicket(ctx context.Context) (ticket, question string, expiresIn int, err error) {
	if s == nil || s.c == nil {
		return "", "", 0, errors.New("captcha store unavailable")
	}
	q, ans := GenerateQuestion()
	ticket = uuid.NewString()
	rec := record{Question: q, Answer: ans, Verified: false}
	s.c.SetJSON(ctx, key(ticket), rec, ticketTTL)
	return ticket, q, int(ticketTTL.Seconds()), nil
}

// Verify 校验答案；成功后将 verified 置 true 并延长可消费窗口。
func (s *Store) Verify(ctx context.Context, ticket string, userAnswer float64) error {
	if s == nil || s.c == nil {
		return errors.New("captcha store unavailable")
	}
	if ticket == "" || math.IsNaN(userAnswer) || math.IsInf(userAnswer, 0) {
		return ErrInvalidAnswer
	}
	var rec record
	if !s.c.GetJSON(ctx, key(ticket), &rec) {
		return ErrTicketNotFound
	}
	if rec.Verified {
		return ErrAlreadyConsumed
	}
	if math.Abs(userAnswer-rec.Answer) > 1e-6 {
		return ErrWrongAnswer
	}
	rec.Verified = true
	s.c.SetJSON(ctx, key(ticket), rec, verifiedConsume)
	return nil
}

// ConsumeVerified 发送短信前消费已验证的 ticket（一次性）。
func (s *Store) ConsumeVerified(ctx context.Context, ticket string) error {
	if s == nil || s.c == nil {
		return errors.New("captcha store unavailable")
	}
	if ticket == "" {
		return ErrTicketNotFound
	}
	var rec record
	k := key(ticket)
	if !s.c.GetJSON(ctx, k, &rec) {
		return ErrTicketNotFound
	}
	if !rec.Verified {
		return ErrNotVerified
	}
	s.c.Delete(ctx, k)
	return nil
}

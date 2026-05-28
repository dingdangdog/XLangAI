package sms

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"xlangai/server/internal/repository"
)

var (
	ErrNotConfigured = errors.New("sms service not configured")
	ErrUnsupported   = errors.New("unsupported sms provider")
)

// Service 按 sys_sms_service_configs 中唯一启用的配置发送验证码短信。
type Service struct {
	repo        *repository.SmsConfigRepo
	verboseLogs bool
}

func NewService(repo *repository.SmsConfigRepo, verboseLogs bool) *Service {
	return &Service{repo: repo, verboseLogs: verboseLogs}
}

// SendVerificationCode 向手机号发送验证码。无启用配置且 verboseLogs 时仅打日志（开发联调）。
func (s *Service) SendVerificationCode(ctx context.Context, phone, code string) error {
	if s == nil || s.repo == nil {
		return ErrNotConfigured
	}
	cfg, err := s.repo.GetActive(ctx)
	if err != nil {
		return err
	}
	if cfg == nil {
		if s.verboseLogs {
			log.Printf("[sms] no active config; OTP for phone=%s code=%s (not sent)\n", phone, code)
			return nil
		}
		return ErrNotConfigured
	}
	phone = strings.TrimSpace(phone)
	code = strings.TrimSpace(code)
	if phone == "" || code == "" {
		return fmt.Errorf("phone and code required")
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "aliyun", "alibaba", "aliyun_sms":
		return sendAliyun(cfg, phone, code)
	case "tencent", "tencentcloud", "qcloud":
		return sendTencent(cfg, phone, code)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupported, cfg.Provider)
	}
}

func configString(cfg map[string]string, key, fallback string) string {
	if cfg == nil {
		return fallback
	}
	if v := strings.TrimSpace(cfg[key]); v != "" {
		return v
	}
	return fallback
}

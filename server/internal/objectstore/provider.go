package objectstore

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CloudBackend 云对象存储厂商实现（local 由 Upload 单独处理）。
type CloudBackend interface {
	ID() string
	CredsReady(cfg *RuntimeConfig) bool
	SupportsDirectUpload() bool
	Upload(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error
	PresignPut(ctx context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (*PresignResult, error)
	Download(ctx context.Context, cfg *RuntimeConfig, key string) ([]byte, error)
}

var cloudBackends = map[string]CloudBackend{
	ProviderCloudflareR2: r2Backend{},
	ProviderQiniu:        qiniuBackend{},
	ProviderAliyunOSS:    aliyunBackend{},
}

func normalizeProvider(p string) string {
	return strings.ToLower(strings.TrimSpace(p))
}

func cloudBackendFor(cfg *RuntimeConfig) (CloudBackend, error) {
	if cfg == nil {
		return nil, ErrNotConfigured
	}
	b, ok := cloudBackends[normalizeProvider(cfg.Provider)]
	if !ok {
		return nil, fmt.Errorf("unsupported object storage provider: %s", cfg.Provider)
	}
	return b, nil
}

// SupportsDirectUploadProvider 判断厂商是否支持客户端直传。
func SupportsDirectUploadProvider(provider string) bool {
	b, ok := cloudBackends[normalizeProvider(provider)]
	if !ok {
		return false
	}
	return b.SupportsDirectUpload()
}

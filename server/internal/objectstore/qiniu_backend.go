package objectstore

import (
	"context"
	"time"
)

type qiniuBackend struct{}

func (qiniuBackend) ID() string { return ProviderQiniu }

func (qiniuBackend) CredsReady(cfg *RuntimeConfig) bool { return qiniuCredsReady(cfg) }

func (qiniuBackend) SupportsDirectUpload() bool { return true }

func (qiniuBackend) Upload(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	return uploadQiniu(ctx, cfg, key, data, contentType)
}

func (b qiniuBackend) PresignPut(ctx context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (*PresignResult, error) {
	return presignQiniuPut(ctx, cfg, key, contentType, ttl)
}

func (qiniuBackend) Download(ctx context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	return downloadQiniu(ctx, cfg, key)
}

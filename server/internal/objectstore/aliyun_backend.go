package objectstore

import (
	"context"
	"time"
)

type aliyunBackend struct{}

func (aliyunBackend) ID() string { return ProviderAliyunOSS }

func (aliyunBackend) CredsReady(cfg *RuntimeConfig) bool { return aliyunCredsReady(cfg) }

func (aliyunBackend) SupportsDirectUpload() bool { return true }

func (aliyunBackend) Upload(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	return uploadAliyun(ctx, cfg, key, data, contentType)
}

func (b aliyunBackend) PresignPut(ctx context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (*PresignResult, error) {
	uploadURL, err := presignAliyunPut(ctx, cfg, key, contentType, ttl)
	if err != nil {
		return nil, err
	}
	return &PresignResult{
		Method:      "PUT",
		UploadURL:   uploadURL,
		PublicURL:   joinPublicURL(cfg.PublicBaseURL, key),
		ObjectKey:   key,
		ContentType: contentType,
		ExpiresAt:   time.Now().UTC().Add(ttl),
		Provider:    b.ID(),
	}, nil
}

func (aliyunBackend) Download(ctx context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	return downloadAliyun(ctx, cfg, key)
}

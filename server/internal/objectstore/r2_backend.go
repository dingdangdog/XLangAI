package objectstore

import (
	"context"
	"time"
)

type r2Backend struct{}

func (r2Backend) ID() string { return ProviderCloudflareR2 }

func (r2Backend) CredsReady(cfg *RuntimeConfig) bool { return cloudCredsReady(cfg) }

func (r2Backend) SupportsDirectUpload() bool { return true }

func (r2Backend) Upload(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	return uploadR2(ctx, cfg, key, data, contentType)
}

func (b r2Backend) PresignPut(ctx context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (*PresignResult, error) {
	uploadURL, err := presignR2Put(ctx, cfg, key, contentType, ttl)
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

func (r2Backend) Download(ctx context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	return downloadR2(ctx, cfg, key)
}

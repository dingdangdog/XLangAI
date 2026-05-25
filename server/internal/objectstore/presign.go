package objectstore

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PresignInput 预签名上传参数。
type PresignInput struct {
	Category    Category
	Ext         string
	ContentType string
	TTL         time.Duration
}

// PresignResult 客户端直传所需字段。
type PresignResult struct {
	Method      string
	UploadURL   string
	PublicURL   string
	ObjectKey   string
	ContentType string
	ExpiresAt   time.Time
}

// PresignPut 为 PutObject 生成预签名 URL（R2 / 阿里云 OSS S3 兼容）。
func PresignPut(ctx context.Context, cfg *RuntimeConfig, in PresignInput) (*PresignResult, error) {
	if cfg == nil {
		return nil, ErrNotConfigured
	}
	ttl := in.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	key := BuildObjectKey(in.Category, in.Ext)
	ct := strings.TrimSpace(in.ContentType)
	if ct == "" {
		ct = "application/octet-stream"
	}

	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case ProviderCloudflareR2:
		if !cloudCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		uploadURL, err := presignR2Put(ctx, cfg, key, ct, ttl)
		if err != nil {
			return nil, err
		}
		return &PresignResult{
			Method:      "PUT",
			UploadURL:   uploadURL,
			PublicURL:   joinPublicURL(cfg.PublicBaseURL, key),
			ObjectKey:   key,
			ContentType: ct,
			ExpiresAt:   time.Now().UTC().Add(ttl),
		}, nil
	case ProviderAliyunOSS:
		if !aliyunCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		uploadURL, err := presignAliyunPut(ctx, cfg, key, ct, ttl)
		if err != nil {
			return nil, err
		}
		return &PresignResult{
			Method:      "PUT",
			UploadURL:   uploadURL,
			PublicURL:   joinPublicURL(cfg.PublicBaseURL, key),
			ObjectKey:   key,
			ContentType: ct,
			ExpiresAt:   time.Now().UTC().Add(ttl),
		}, nil
	default:
		return nil, fmt.Errorf("presigned upload not supported for provider %s", provider)
	}
}

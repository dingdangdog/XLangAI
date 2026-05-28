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
	// UploadToken 七牛云表单上传 token（method=POST 时使用）。
	UploadToken string
	Provider    string
}

// PresignPut 为客户端直传生成凭证（R2/OSS 预签名 PUT；七牛 upload token + POST）。
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

	backend, err := cloudBackendFor(cfg)
	if err != nil {
		return nil, err
	}
	if !backend.CredsReady(cfg) {
		return nil, ErrNotConfigured
	}
	if !backend.SupportsDirectUpload() {
		return nil, fmt.Errorf("presigned upload not supported for provider %s", cfg.Provider)
	}
	res, err := backend.PresignPut(ctx, cfg, key, ct, ttl)
	if err != nil {
		return nil, err
	}
	if res.ContentType == "" {
		res.ContentType = ct
	}
	return res, nil
}

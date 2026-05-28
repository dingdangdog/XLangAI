package objectstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"xlangai/server/internal/storage"
)

var ErrNotConfigured = errors.New("object storage not configured")

// LocalDirs 本地落盘目录（provider=local 或无 active 云配置时使用）。
type LocalDirs struct {
	AvatarDir string
	AudioDir  string
}

// UploadInput 上传参数。
type UploadInput struct {
	Category    Category
	Data        []byte
	Ext         string
	ContentType string
}

// UploadResult 返回可写入 DB 的访问 URL。
type UploadResult struct {
	URL      string
	Filename string // 仅本地模式有值，供 Serve* 路由使用
}

// Upload 按配置上传到云或本地。
func Upload(ctx context.Context, cfg *RuntimeConfig, local LocalDirs, in UploadInput) (*UploadResult, error) {
	ext := normalizeExt(in.Ext)
	key := BuildObjectKey(in.Category, ext)

	provider := ProviderLocal
	if cfg != nil && strings.TrimSpace(cfg.Provider) != "" {
		provider = normalizeProvider(cfg.Provider)
	}

	if provider == ProviderLocal {
		return uploadLocal(local, in.Category, in.Data, ext)
	}

	backend, err := cloudBackendFor(cfg)
	if err != nil {
		return nil, err
	}
	if !backend.CredsReady(cfg) {
		return nil, ErrNotConfigured
	}
	if err := backend.Upload(ctx, cfg, key, in.Data, in.ContentType); err != nil {
		return nil, err
	}
	return &UploadResult{URL: joinPublicURL(cfg.PublicBaseURL, key)}, nil
}

func uploadLocal(local LocalDirs, cat Category, data []byte, ext string) (*UploadResult, error) {
	dir := local.AudioDir
	prefix := "/api/v1/audio/"
	if cat == CategoryAvatar {
		dir = local.AvatarDir
		prefix = "/api/v1/avatars/"
	}
	if dir == "" {
		if cat == CategoryAvatar {
			dir = "./storage/avatars"
		} else {
			dir = "./storage/audio"
		}
	}
	name, err := storage.SaveAudioWithExt(dir, data, ext)
	if err != nil {
		return nil, err
	}
	return &UploadResult{URL: prefix + name, Filename: name}, nil
}

func normalizeExt(ext string) string {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func cloudCredsReady(cfg *RuntimeConfig) bool {
	return cfg.Endpoint != "" && cfg.AccessKey != "" && cfg.SecretKey != "" && cfg.Bucket != "" && cfg.PublicBaseURL != ""
}

func qiniuCredsReady(cfg *RuntimeConfig) bool {
	return cfg.AccessKey != "" && cfg.SecretKey != "" && cfg.Bucket != "" && cfg.PublicBaseURL != ""
}

func aliyunCredsReady(cfg *RuntimeConfig) bool {
	return cfg.Endpoint != "" && cfg.AccessKey != "" && cfg.SecretKey != "" && cfg.Bucket != "" && cfg.PublicBaseURL != ""
}

// DeleteObject 从云存储删除对象（供 Manager 等复用同一实现时参考）。
func DeleteObject(ctx context.Context, cfg *RuntimeConfig, key string) error {
	if cfg == nil || strings.TrimSpace(key) == "" {
		return ErrNotConfigured
	}
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	switch normalizeProvider(cfg.Provider) {
	case ProviderCloudflareR2:
		if !cloudCredsReady(cfg) {
			return ErrNotConfigured
		}
		return deleteR2(ctx, cfg, key)
	case ProviderAliyunOSS:
		if !aliyunCredsReady(cfg) {
			return ErrNotConfigured
		}
		return deleteAliyun(ctx, cfg, key)
	case ProviderQiniu:
		if !qiniuCredsReady(cfg) {
			return ErrNotConfigured
		}
		return deleteQiniu(ctx, cfg, key)
	default:
		return fmt.Errorf("delete not supported for provider %s", cfg.Provider)
	}
}

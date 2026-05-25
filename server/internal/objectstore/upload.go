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
		provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	}

	switch provider {
	case ProviderLocal:
		return uploadLocal(local, in.Category, in.Data, ext)
	case ProviderCloudflareR2:
		if cfg == nil || !cloudCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		if err := uploadR2(ctx, cfg, key, in.Data, in.ContentType); err != nil {
			return nil, err
		}
		return &UploadResult{URL: joinPublicURL(cfg.PublicBaseURL, key)}, nil
	case ProviderQiniu:
		if cfg == nil || !qiniuCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		if err := uploadQiniu(ctx, cfg, key, in.Data, in.ContentType); err != nil {
			return nil, err
		}
		return &UploadResult{URL: joinPublicURL(cfg.PublicBaseURL, key)}, nil
	case ProviderAliyunOSS:
		if cfg == nil || !aliyunCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		if err := uploadAliyun(ctx, cfg, key, in.Data, in.ContentType); err != nil {
			return nil, err
		}
		return &UploadResult{URL: joinPublicURL(cfg.PublicBaseURL, key)}, nil
	default:
		return nil, fmt.Errorf("unsupported object storage provider: %s", provider)
	}
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

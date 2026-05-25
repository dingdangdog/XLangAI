package objectstore

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// DownloadObject 从云存储读取对象字节（供 STT 等服务端处理使用）。
func DownloadObject(ctx context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	if cfg == nil || strings.TrimSpace(key) == "" {
		return nil, ErrNotConfigured
	}
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderCloudflareR2:
		if !cloudCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		return downloadR2(ctx, cfg, key)
	case ProviderAliyunOSS:
		if !aliyunCredsReady(cfg) {
			return nil, ErrNotConfigured
		}
		return downloadAliyun(ctx, cfg, key)
	default:
		return nil, fmt.Errorf("download not supported for provider %s", cfg.Provider)
	}
}

// DownloadByPublicURL 根据公网 URL 解析 key 并下载。
func DownloadByPublicURL(ctx context.Context, cfg *RuntimeConfig, objectURL string) ([]byte, error) {
	key := KeyFromPublicURL(cfg.PublicBaseURL, objectURL)
	if key == "" {
		return nil, fmt.Errorf("object url does not match configured public base")
	}
	return DownloadObject(ctx, cfg, key)
}

func readAll(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	return io.ReadAll(r)
}

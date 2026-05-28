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
	backend, err := cloudBackendFor(cfg)
	if err != nil {
		return nil, err
	}
	if !backend.CredsReady(cfg) {
		return nil, ErrNotConfigured
	}
	return backend.Download(ctx, cfg, key)
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

package objectstore

import (
	"encoding/json"
	"strings"

	"xlangai/server/internal/repository"
)

const (
	ProviderLocal        = "local"
	ProviderCloudflareR2 = "cloudflare_r2"
	ProviderQiniu        = "qiniu"
	ProviderAliyunOSS    = "aliyun_oss"
)

// RuntimeConfig 由数据库 active 配置解析而来。
type RuntimeConfig struct {
	Provider      string
	Endpoint      string
	PublicBaseURL string
	AccessKey     string
	SecretKey     string
	Bucket        string
	Region        string
	Extra         map[string]string
}

func FromRepo(cfg *repository.ObjectStorageConfig) *RuntimeConfig {
	if cfg == nil {
		return nil
	}
	rc := &RuntimeConfig{
		Provider:      strings.ToLower(strings.TrimSpace(cfg.Provider)),
		Endpoint:      cfg.BaseURL,
		PublicBaseURL: cfg.PublicBaseURL,
		AccessKey:     cfg.APIKey,
		SecretKey:     cfg.SecretKey,
		Bucket:        cfg.Bucket,
		Region:        cfg.Region,
		Extra:         cfg.Config,
	}
	if rc.Extra == nil {
		rc.Extra = map[string]string{}
	}
	return rc
}

func ParseExtraJSON(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil || m == nil {
		return map[string]string{}
	}
	return m
}

func joinPublicURL(base, key string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if base == "" {
		return key
	}
	if key == "" {
		return base
	}
	if strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://") {
		return base + "/" + key
	}
	return "https://" + base + "/" + key
}

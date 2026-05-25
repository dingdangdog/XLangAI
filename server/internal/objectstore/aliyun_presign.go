package objectstore

import (
	"context"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func presignAliyunPut(_ context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (string, error) {
	client, err := oss.New(strings.TrimSpace(cfg.Endpoint), cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return "", err
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return "", err
	}
	opts := []oss.Option{}
	if strings.TrimSpace(contentType) != "" {
		opts = append(opts, oss.ContentType(contentType))
	}
	sec := int64(ttl.Seconds())
	if sec < 60 {
		sec = 60
	}
	return bucket.SignURL(key, oss.HTTPPut, sec, opts...)
}

func downloadAliyun(_ context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	client, err := oss.New(strings.TrimSpace(cfg.Endpoint), cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, err
	}
	body, err := bucket.GetObject(key)
	if err != nil {
		return nil, err
	}
	return readAll(body)
}

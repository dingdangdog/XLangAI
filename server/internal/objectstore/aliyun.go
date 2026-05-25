package objectstore

import (
	"bytes"
	"context"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func uploadAliyun(_ context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	client, err := oss.New(strings.TrimSpace(cfg.Endpoint), cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return err
	}
	opts := []oss.Option{}
	ct := strings.TrimSpace(contentType)
	if ct != "" {
		opts = append(opts, oss.ContentType(ct))
	}
	return bucket.PutObject(key, bytes.NewReader(data), opts...)
}

package objectstore

import (
	"bytes"
	"context"
	"strings"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

func uploadQiniu(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	putPolicy := storage.PutPolicy{Scope: cfg.Bucket}
	upToken := putPolicy.UploadToken(mac)
	formUploader := storage.NewFormUploader(&storage.Config{})
	ret := storage.PutRet{}
	extra := storage.PutExtra{}
	ct := strings.TrimSpace(contentType)
	if ct != "" {
		extra.MimeType = ct
	}
	return formUploader.Put(ctx, &ret, upToken, key, bytes.NewReader(data), int64(len(data)), &extra)
}

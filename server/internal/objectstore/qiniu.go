package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

func qiniuStorageConfig(cfg *RuntimeConfig) *storage.Config {
	sc := &storage.Config{}
	zone := strings.ToLower(strings.TrimSpace(cfg.Extra["zone"]))
	switch zone {
	case "z1":
		sc.Zone = &storage.ZoneHuabei
	case "z2":
		sc.Zone = &storage.ZoneHuanan
	case "na0":
		sc.Zone = &storage.ZoneBeimei
	case "as0":
		sc.Zone = &storage.ZoneXinjiapo
	default:
		sc.Zone = &storage.ZoneHuadong
	}
	if endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"); endpoint != "" {
		sc.UseHTTPS = strings.HasPrefix(endpoint, "https://")
		sc.UseCdnDomains = false
	}
	return sc
}

func qiniuUploadHost(cfg *RuntimeConfig) string {
	if endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"); endpoint != "" {
		return endpoint
	}
	zone := strings.ToLower(strings.TrimSpace(cfg.Extra["zone"]))
	switch zone {
	case "z1":
		return "https://upload-z1.qiniup.com"
	case "z2":
		return "https://upload-z2.qiniup.com"
	case "na0":
		return "https://upload-na0.qiniup.com"
	case "as0":
		return "https://upload-as0.qiniup.com"
	default:
		return "https://upload.qiniup.com"
	}
}

func uploadQiniu(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	putPolicy := storage.PutPolicy{Scope: cfg.Bucket + ":" + key}
	upToken := putPolicy.UploadToken(mac)
	formUploader := storage.NewFormUploader(qiniuStorageConfig(cfg))
	ret := storage.PutRet{}
	extra := storage.PutExtra{}
	ct := strings.TrimSpace(contentType)
	if ct != "" {
		extra.MimeType = ct
	}
	return formUploader.Put(ctx, &ret, upToken, key, bytes.NewReader(data), int64(len(data)), &extra)
}

func presignQiniuPut(_ context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (*PresignResult, error) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	putPolicy := storage.PutPolicy{
		Scope:   cfg.Bucket + ":" + key,
		Expires: uint64(time.Now().Add(ttl).Unix()),
	}
	ct := strings.TrimSpace(contentType)
	if ct != "" {
		putPolicy.MimeLimit = ct
	}
	token := putPolicy.UploadToken(mac)
	return &PresignResult{
		Method:      "POST",
		UploadURL:   qiniuUploadHost(cfg),
		UploadToken: token,
		PublicURL:   joinPublicURL(cfg.PublicBaseURL, key),
		ObjectKey:   key,
		ContentType: ct,
		ExpiresAt:   time.Now().UTC().Add(ttl),
		Provider:    ProviderQiniu,
	}, nil
}

func downloadQiniu(_ context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	bm := storage.NewBucketManager(mac, qiniuStorageConfig(cfg))
	out, err := bm.Get(cfg.Bucket, key, nil)
	if err != nil {
		return nil, err
	}
	if out == nil || out.Body == nil {
		return nil, fmt.Errorf("qiniu get object returned empty body")
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}

func deleteQiniu(_ context.Context, cfg *RuntimeConfig, key string) error {
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	bm := storage.NewBucketManager(mac, qiniuStorageConfig(cfg))
	return bm.Delete(cfg.Bucket, key)
}

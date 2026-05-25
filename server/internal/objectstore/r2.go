package objectstore

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func newR2Client(cfg *RuntimeConfig) *s3.Client {
	endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	cred := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	awsCfg := aws.Config{
		Region:      "auto",
		Credentials: cred,
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

func uploadR2(ctx context.Context, cfg *RuntimeConfig, key string, data []byte, contentType string) error {
	client := newR2Client(cfg)
	ct := strings.TrimSpace(contentType)
	if ct == "" {
		ct = "application/octet-stream"
	}
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(ct),
	})
	return err
}

func presignR2Put(ctx context.Context, cfg *RuntimeConfig, key, contentType string, ttl time.Duration) (string, error) {
	client := newR2Client(cfg)
	presigner := s3.NewPresignClient(client)
	out, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func downloadR2(ctx context.Context, cfg *RuntimeConfig, key string) ([]byte, error) {
	client := newR2Client(cfg)
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return readAll(out.Body)
}

package redis

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// New 从 REDIS_URL 创建客户端；URL 为空则返回 nil（调用方做 nil 判断）。
func New(redisURL string) *redis.Client {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return nil
	}
	var opts *redis.Options
	var err error
	if strings.HasPrefix(redisURL, "redis://") || strings.HasPrefix(redisURL, "rediss://") {
		opts, err = redis.ParseURL(redisURL)
	} else {
		opts = &redis.Options{Addr: redisURL}
	}
	if err != nil {
		log.Printf("redis: 解析 REDIS_URL 失败: %v，将不使用缓存", err)
		return nil
	}
	c := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		log.Printf("redis: Ping 失败: %v，将不使用缓存", err)
		_ = c.Close()
		return nil
	}
	log.Printf("redis: 已连接")
	return c
}

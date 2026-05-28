package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// BackendKind 当前缓存后端类型。
type BackendKind string

const (
	BackendRedis  BackendKind = "redis"
	BackendMemory BackendKind = "memory"
)

type backend interface {
	kind() BackendKind
	get(ctx context.Context, key string) (string, bool)
	set(ctx context.Context, key, val string, ttl time.Duration) bool
	delete(ctx context.Context, keys ...string)
}

type redisBackend struct {
	rdb *redis.Client
}

func (b *redisBackend) kind() BackendKind { return BackendRedis }

func (b *redisBackend) get(ctx context.Context, key string) (string, bool) {
	s, err := b.rdb.Get(ctx, key).Result()
	if err != nil || s == "" {
		return "", false
	}
	return s, true
}

func (b *redisBackend) set(ctx context.Context, key, val string, ttl time.Duration) bool {
	return b.rdb.Set(ctx, key, val, ttl).Err() == nil
}

func (b *redisBackend) delete(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	_ = b.rdb.Del(ctx, keys...).Err()
}

package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var defaultCache *Cache

// Init 创建并注册进程全局缓存实例；main 启动时调用一次。
// 已连接 Redis 时使用 Redis，否则自动降级为进程内内存缓存。
func Init(rdb *redis.Client) *Cache {
	c := New(rdb)
	defaultCache = c
	return c
}

// Default 返回 Init 注册的全局实例（未 Init 时为 nil）。
func Default() *Cache {
	return defaultCache
}

// BackendKindOf 返回全局缓存后端类型；未 Init 时返回 BackendMemory。
func BackendKindOf() BackendKind {
	if defaultCache == nil {
		return BackendMemory
	}
	return defaultCache.BackendKind()
}

// HasRedis 全局实例是否使用 Redis 后端。
func HasRedis() bool {
	return defaultCache != nil && defaultCache.HasRedis()
}

func GetJSON(ctx context.Context, key string, dest interface{}) bool {
	if defaultCache == nil {
		return false
	}
	return defaultCache.GetJSON(ctx, key, dest)
}

func SetJSON(ctx context.Context, key string, v interface{}, ttl time.Duration) {
	if defaultCache == nil {
		return
	}
	defaultCache.SetJSON(ctx, key, v, ttl)
}

func Delete(ctx context.Context, keys ...string) {
	if defaultCache == nil {
		return
	}
	defaultCache.Delete(ctx, keys...)
}

func SetPlain(ctx context.Context, key, val string, ttl time.Duration) bool {
	if defaultCache == nil {
		return false
	}
	return defaultCache.SetPlain(ctx, key, val, ttl)
}

func GetPlain(ctx context.Context, key string) (string, bool) {
	if defaultCache == nil {
		return "", false
	}
	return defaultCache.GetPlain(ctx, key)
}

// GetOrSetJSON 读缓存；未命中时执行 loader 并写入（泛型便于业务侧直接使用）。
func GetOrSetJSON[T any](ctx context.Context, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var zero T
	if defaultCache != nil {
		var cached T
		if defaultCache.GetJSON(ctx, key, &cached) {
			return cached, nil
		}
	}
	v, err := loader()
	if err != nil {
		return zero, err
	}
	if defaultCache != nil {
		defaultCache.SetJSON(ctx, key, v, ttl)
	}
	return v, nil
}

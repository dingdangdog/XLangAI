package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	keyLangList     = "xlangai:languages:v1"
	keyPrincipalFmt = "xlangai:principal:v1:%s"
)

// Cache 统一缓存门面：自动选择 Redis 或进程内内存后端。
type Cache struct {
	backend backend
}

// New 创建缓存实例。rdb 非 nil 且可用时使用 Redis，否则使用进程内内存。
func New(rdb *redis.Client) *Cache {
	if rdb != nil {
		log.Printf("cache: 使用 Redis 后端")
		return &Cache{backend: &redisBackend{rdb: rdb}}
	}
	log.Printf("cache: 使用进程内内存后端（单实例可用；多实例部署请配置 REDIS_URL）")
	return &Cache{backend: newMemoryBackend()}
}

// BackendKind 返回当前后端类型。
func (c *Cache) BackendKind() BackendKind {
	if c == nil || c.backend == nil {
		return BackendMemory
	}
	return c.backend.kind()
}

// HasRedis 是否使用 Redis 后端。
func (c *Cache) HasRedis() bool {
	return c != nil && c.backend != nil && c.backend.kind() == BackendRedis
}

func (c *Cache) enabled() bool {
	return c != nil && c.backend != nil
}

func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) bool {
	if !c.enabled() {
		return false
	}
	s, ok := c.backend.get(ctx, key)
	if !ok || s == "" {
		return false
	}
	if json.Unmarshal([]byte(s), dest) != nil {
		return false
	}
	return true
}

func (c *Cache) SetJSON(ctx context.Context, key string, v interface{}, ttl time.Duration) {
	if !c.enabled() {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	_ = c.backend.set(ctx, key, string(b), ttl)
}

func (c *Cache) Delete(ctx context.Context, keys ...string) {
	if !c.enabled() {
		return
	}
	c.backend.delete(ctx, keys...)
}

// SetPlain 写入原始字符串（非 JSON），TTL 到期自动删除。
func (c *Cache) SetPlain(ctx context.Context, key, val string, ttl time.Duration) bool {
	if !c.enabled() {
		return false
	}
	return c.backend.set(ctx, key, val, ttl)
}

// GetPlain 读取 SetPlain 写入的值。
func (c *Cache) GetPlain(ctx context.Context, key string) (string, bool) {
	if !c.enabled() {
		return "", false
	}
	return c.backend.get(ctx, key)
}

func LangListKey() string { return keyLangList }

func PrincipalKey(userID string) string {
	return fmt.Sprintf(keyPrincipalFmt, userID)
}

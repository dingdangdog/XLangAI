package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	keyLangList     = "xlangai:languages:v1"
	keyPrincipalFmt = "xlangai:principal:v1:%s"
)

type Cache struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) enabled() bool {
	return c != nil && c.rdb != nil
}

// HasRedis 是否已连接 Redis（SetPlain/GetPlain 等可用）。
func (c *Cache) HasRedis() bool {
	return c.enabled()
}

func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) bool {
	if !c.enabled() {
		return false
	}
	s, err := c.rdb.Get(ctx, key).Result()
	if err != nil || s == "" {
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
	_ = c.rdb.Set(ctx, key, b, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) {
	if !c.enabled() || len(keys) == 0 {
		return
	}
	_ = c.rdb.Del(ctx, keys...).Err()
}

// SetPlain 写入原始字符串（非 JSON），TTL 到期自动删除。
func (c *Cache) SetPlain(ctx context.Context, key, val string, ttl time.Duration) bool {
	if !c.enabled() {
		return false
	}
	return c.rdb.Set(ctx, key, val, ttl).Err() == nil
}

// GetPlain 读取 SetPlain 写入的值。
func (c *Cache) GetPlain(ctx context.Context, key string) (string, bool) {
	if !c.enabled() {
		return "", false
	}
	s, err := c.rdb.Get(ctx, key).Result()
	if err != nil || s == "" {
		return "", false
	}
	return s, true
}

func LangListKey() string { return keyLangList }

func PrincipalKey(userID string) string {
	return fmt.Sprintf(keyPrincipalFmt, userID)
}

package cache

import (
	"context"
	"sync"
	"time"
)

type memEntry struct {
	value     string
	expiresAt time.Time
}

type memoryBackend struct {
	mu    sync.RWMutex
	items map[string]memEntry
	stop  chan struct{}
}

func newMemoryBackend() *memoryBackend {
	m := &memoryBackend{
		items: make(map[string]memEntry),
		stop:  make(chan struct{}),
	}
	go m.cleanupLoop()
	return m
}

func (b *memoryBackend) kind() BackendKind { return BackendMemory }

func (b *memoryBackend) get(_ context.Context, key string) (string, bool) {
	b.mu.RLock()
	e, ok := b.items[key]
	b.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		b.mu.Lock()
		delete(b.items, key)
		b.mu.Unlock()
		return "", false
	}
	return e.value, true
}

func (b *memoryBackend) set(_ context.Context, key, val string, ttl time.Duration) bool {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	b.mu.Lock()
	b.items[key] = memEntry{value: val, expiresAt: exp}
	b.mu.Unlock()
	return true
}

func (b *memoryBackend) delete(_ context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	b.mu.Lock()
	for _, key := range keys {
		delete(b.items, key)
	}
	b.mu.Unlock()
}

func (b *memoryBackend) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.evictExpired()
		case <-b.stop:
			return
		}
	}
}

func (b *memoryBackend) evictExpired() {
	now := time.Now()
	b.mu.Lock()
	for key, e := range b.items {
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			delete(b.items, key)
		}
	}
	b.mu.Unlock()
}

func (b *memoryBackend) close() {
	select {
	case <-b.stop:
	default:
		close(b.stop)
	}
}

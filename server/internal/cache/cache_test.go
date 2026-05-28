package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryBackendPlainAndJSON(t *testing.T) {
	c := New(nil)
	if c.BackendKind() != BackendMemory {
		t.Fatalf("expected memory backend, got %s", c.BackendKind())
	}

	ctx := context.Background()
	if !c.SetPlain(ctx, "k1", "v1", time.Minute) {
		t.Fatal("SetPlain failed")
	}
	if v, ok := c.GetPlain(ctx, "k1"); !ok || v != "v1" {
		t.Fatalf("GetPlain = %q, %v", v, ok)
	}

	type item struct {
		Name string `json:"name"`
	}
	c.SetJSON(ctx, "j1", item{Name: "test"}, time.Minute)
	var got item
	if !c.GetJSON(ctx, "j1", &got) || got.Name != "test" {
		t.Fatalf("GetJSON = %+v", got)
	}

	c.Delete(ctx, "k1", "j1")
	if _, ok := c.GetPlain(ctx, "k1"); ok {
		t.Fatal("key should be deleted")
	}
}

func TestMemoryBackendTTL(t *testing.T) {
	c := New(nil)
	ctx := context.Background()
	_ = c.SetPlain(ctx, "exp", "x", 20*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	if _, ok := c.GetPlain(ctx, "exp"); ok {
		t.Fatal("expired key should miss")
	}
}

func TestGlobalInit(t *testing.T) {
	Init(nil)
	if Default() == nil {
		t.Fatal("Default() nil after Init")
	}
	if BackendKindOf() != BackendMemory {
		t.Fatalf("BackendKindOf = %s", BackendKindOf())
	}

	ctx := context.Background()
	SetPlain(ctx, "g1", "ok", time.Minute)
	if v, ok := GetPlain(ctx, "g1"); !ok || v != "ok" {
		t.Fatalf("global GetPlain = %q, %v", v, ok)
	}

	type payload struct {
		N int `json:"n"`
	}
	v, err := GetOrSetJSON(ctx, "g2", time.Minute, func() (payload, error) {
		return payload{N: 42}, nil
	})
	if err != nil || v.N != 42 {
		t.Fatalf("GetOrSetJSON first = %+v, err=%v", v, err)
	}
	v2, err := GetOrSetJSON(ctx, "g2", time.Minute, func() (payload, error) {
		t.Fatal("loader should not run on cache hit")
		return payload{}, nil
	})
	if err != nil || v2.N != 42 {
		t.Fatalf("GetOrSetJSON hit = %+v, err=%v", v2, err)
	}
}

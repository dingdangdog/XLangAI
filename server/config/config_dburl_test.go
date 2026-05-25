package config

import (
	"strings"
	"testing"
)

func TestNormalizeDBURL_stripsPrismaSchema(t *testing.T) {
	raw := "postgresql://u:p@host:5432/xlangai?schema=public&sslmode=disable"
	got := normalizeDBURL(raw)
	if got == raw {
		t.Fatalf("expected schema stripped, got %q", got)
	}
	if strings.Contains(got, "schema=") {
		t.Fatalf("schema still present: %q", got)
	}
	if !strings.Contains(got, "sslmode=disable") {
		t.Fatalf("sslmode should remain: %q", got)
	}
}

func TestNormalizeDBURL_invalidURLUnchanged(t *testing.T) {
	raw := "not-a-url"
	if got := normalizeDBURL(raw); got != raw {
		t.Fatalf("got %q", got)
	}
}

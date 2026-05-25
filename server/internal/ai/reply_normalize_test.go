package ai

import (
	"strings"
	"testing"
)

func TestNormalizeAssistantReply_stripsEmoji(t *testing.T) {
	in := "Nice 👍 to meet you! ✅ All good."
	out := NormalizeAssistantReply(in)
	if out == "" {
		t.Fatal("empty output")
	}
	for _, r := range []string{"👍", "✅"} {
		if strings.Contains(out, r) {
			t.Fatalf("still contains %q: %q", r, out)
		}
	}
}

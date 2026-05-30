package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBundledAudioDirDevFallback(t *testing.T) {
	t.Chdir(filepath.Join("..", "..", ".."))
	t.Setenv("BUNDLED_AUDIO_DIR", "")

	resolvedBundledAudioDirLoaded = false
	resolvedBundledAudioDir = ""

	dir := resolveBundledAudioDir()
	if !isExistingDir(dir) {
		t.Fatalf("expected existing bundled audio dir, got %q", dir)
	}

	info, err := os.Stat(filepath.Join(dir, "en-US-AvaMultilingualNeural-Ava.mp3"))
	if err != nil {
		t.Fatalf("expected bundled preview mp3 in %q: %v", dir, err)
	}
	if info.Size() == 0 {
		t.Fatalf("bundled preview mp3 is empty")
	}
}

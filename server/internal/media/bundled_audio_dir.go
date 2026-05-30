package media

import (
	"os"
	"path/filepath"
	"strings"
)

const dockerBundledAudioDir = "/app/bootstrap-storage/audio"

var (
	resolvedBundledAudioDir       string
	resolvedBundledAudioDirLoaded bool
)

// resolveBundledAudioDir 返回内置试听音频目录。
// 生产 Docker 镜像通过 BUNDLED_AUDIO_DIR 或 /app/bootstrap-storage/audio 提供；
// 本地开发在未显式配置时自动探测 manager/storage/audio，不改变已配置的生产路径。
func resolveBundledAudioDir() string {
	if resolvedBundledAudioDirLoaded {
		return resolvedBundledAudioDir
	}
	resolvedBundledAudioDirLoaded = true

	if v := strings.TrimSpace(os.Getenv("BUNDLED_AUDIO_DIR")); v != "" {
		resolvedBundledAudioDir = v
		return resolvedBundledAudioDir
	}

	if isExistingDir(dockerBundledAudioDir) {
		resolvedBundledAudioDir = dockerBundledAudioDir
		return resolvedBundledAudioDir
	}

	for _, candidate := range devBundledAudioCandidates() {
		if isExistingDir(candidate) {
			resolvedBundledAudioDir = candidate
			return resolvedBundledAudioDir
		}
	}

	resolvedBundledAudioDir = dockerBundledAudioDir
	return resolvedBundledAudioDir
}

func isExistingDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func devBundledAudioCandidates() []string {
	relPaths := []string{
		filepath.Join("..", "manager", "storage", "audio"),
		filepath.Join("manager", "storage", "audio"),
		filepath.Join("..", "..", "manager", "storage", "audio"),
	}

	var bases []string
	if wd, err := os.Getwd(); err == nil && wd != "" {
		bases = append(bases, wd, filepath.Dir(wd))
	}
	if exe, err := os.Executable(); err == nil {
		bases = append(bases, filepath.Dir(exe))
	}

	seen := make(map[string]struct{}, len(bases)*len(relPaths))
	out := make([]string, 0, len(bases)*len(relPaths))
	for _, base := range bases {
		for _, rel := range relPaths {
			abs, err := filepath.Abs(filepath.Join(base, rel))
			if err != nil {
				continue
			}
			if _, ok := seen[abs]; ok {
				continue
			}
			seen[abs] = struct{}{}
			out = append(out, abs)
		}
	}
	return out
}

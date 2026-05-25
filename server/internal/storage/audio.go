package storage

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// SaveAudio 保存音频到 dir，默认使用 mp3 扩展名。
func SaveAudio(dir string, data []byte) (string, error) {
	return SaveAudioWithExt(dir, data, ".mp3")
}

// SaveAudioWithExt 保存音频到 dir，并保留调用方指定的扩展名。
func SaveAudioWithExt(dir string, data []byte, ext string) (string, error) {
	if dir == "" {
		dir = "./storage/audio"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ext = strings.TrimSpace(ext)
	if ext == "" {
		ext = ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	name := uuid.New().String() + ext
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return "", err
	}
	return name, nil
}

// ResolvePath 返回完整路径。
func ResolvePath(dir, name string) string {
	if dir == "" {
		dir = "./storage/audio"
	}
	return filepath.Join(dir, name)
}

// ReadAudio 读取音频文件。
func ReadAudio(dir, name string) ([]byte, error) {
	p := ResolvePath(dir, name)
	return os.ReadFile(p)
}

package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// TranscodeToAzureSTTWavPCM16k 将任意 ffmpeg 可解码的音频转为 16kHz 单声道 16-bit PCM WAV（Azure 短音频 REST 常用格式）。
func TranscodeToAzureSTTWavPCM16k(ctx context.Context, ffmpegPath string, audio []byte) ([]byte, error) {
	bin := strings.TrimSpace(ffmpegPath)
	if bin == "" {
		bin = "ffmpeg"
	}
	cmd := exec.CommandContext(ctx, bin,
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-f", "wav",
		"-acodec", "pcm_s16le",
		"-ac", "1",
		"-ar", "16000",
		"pipe:1",
	)
	cmd.Stdin = bytes.NewReader(audio)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrFFmpegNotFound
		}
		return nil, fmt.Errorf("ffmpeg transcode: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("ffmpeg transcode produced empty output")
	}
	return out.Bytes(), nil
}

// ErrFFmpegNotFound 表示未找到 ffmpeg 可执行文件（PATH 或 XLANGAI_FFMPEG_PATH）。
var ErrFFmpegNotFound = errors.New("ffmpeg not found: install ffmpeg or set XLANGAI_FFMPEG_PATH")

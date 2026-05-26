package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// TTSLoudnessDefaults 为助手 TTS 响度归一化目标（偏手机外放，比播客 -16 LUFS 略响）。
const (
	TTSLoudnessDefaultI   = -14.0 // 综合响度 LUFS
	TTSLoudnessDefaultTP  = -1.0  // 真峰值上限 dBTP
	TTSLoudnessDefaultLRA = 8.0   // 响度范围
)

// TTSLoudnessOptions 控制 ffmpeg loudnorm 参数；零值字段使用 TTSLoudnessDefault*。
type TTSLoudnessOptions struct {
	TargetLUFS float64
	TruePeakDB float64
	LRA        float64
}

func (o *TTSLoudnessOptions) resolved() (i, tp, lra float64) {
	i, tp, lra = TTSLoudnessDefaultI, TTSLoudnessDefaultTP, TTSLoudnessDefaultLRA
	if o == nil {
		return i, tp, lra
	}
	if o.TargetLUFS != 0 {
		i = o.TargetLUFS
	}
	if o.TruePeakDB != 0 {
		tp = o.TruePeakDB
	}
	if o.LRA != 0 {
		lra = o.LRA
	}
	return i, tp, lra
}

// NormalizeTTSLoudness 将 TTS 字节流归一化到目标响度，缓解各厂商默认音量不一致。
// 成功时返回处理后的音频与 MIME；失败时返回 ErrFFmpegNotFound 或其它错误（调用方应回退为原始音频）。
func NormalizeTTSLoudness(ctx context.Context, ffmpegPath string, audio []byte, mime string, opts *TTSLoudnessOptions) ([]byte, string, error) {
	if len(audio) == 0 {
		return nil, "", fmt.Errorf("tts loudness: empty audio")
	}
	bin := strings.TrimSpace(ffmpegPath)
	if bin == "" {
		bin = "ffmpeg"
	}
	i, tp, lra := opts.resolved()
	outFmt, codecArgs, outMime := ttsLoudnessOutputArgs(mime)
	// 高通去直流 + 预增益 + loudnorm + 限幅，兼顾极轻 TTS 源与峰值保护。
	filter := fmt.Sprintf(
		"highpass=f=80,volume=4dB,loudnorm=I=%g:TP=%g:LRA=%g:print_format=none,alimiter=limit=0.97",
		i, tp, lra,
	)

	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-af", filter,
	}
	args = append(args, codecArgs...)
	args = append(args, "-f", outFmt, "pipe:1")

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = bytes.NewReader(audio)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, "", ErrFFmpegNotFound
		}
		return nil, "", fmt.Errorf("tts loudness ffmpeg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if out.Len() == 0 {
		return nil, "", fmt.Errorf("tts loudness: ffmpeg produced empty output")
	}
	return out.Bytes(), outMime, nil
}

// BoostTTSVolumeFallback 在 loudnorm 失败时的简易增益（仍依赖 ffmpeg）。
func BoostTTSVolumeFallback(ctx context.Context, ffmpegPath string, audio []byte, mime string) ([]byte, string, error) {
	if len(audio) == 0 {
		return nil, "", fmt.Errorf("tts boost: empty audio")
	}
	bin := strings.TrimSpace(ffmpegPath)
	if bin == "" {
		bin = "ffmpeg"
	}
	outFmt, codecArgs, outMime := ttsLoudnessOutputArgs(mime)
	filter := "highpass=f=80,volume=12dB,alimiter=limit=0.97"
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-af", filter,
	}
	args = append(args, codecArgs...)
	args = append(args, "-f", outFmt, "pipe:1")

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = bytes.NewReader(audio)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, "", ErrFFmpegNotFound
		}
		return nil, "", fmt.Errorf("tts boost ffmpeg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if out.Len() == 0 {
		return nil, "", fmt.Errorf("tts boost: ffmpeg produced empty output")
	}
	return out.Bytes(), outMime, nil
}

// ttsLoudnessOutputArgs 按输入 MIME 选择输出容器与编码，尽量保持与原 TTS 一致。
func ttsLoudnessOutputArgs(mime string) (format string, codecArgs []string, outMime string) {
	m := strings.ToLower(strings.TrimSpace(mime))
	switch {
	case strings.Contains(m, "wav"):
		return "wav", []string{"-acodec", "pcm_s16le"}, "audio/wav"
	case strings.Contains(m, "ogg"):
		return "ogg", []string{"-c:a", "libvorbis"}, "audio/ogg"
	default:
		return "mp3", []string{"-c:a", "libmp3lame", "-q:a", "2"}, "audio/mpeg"
	}
}

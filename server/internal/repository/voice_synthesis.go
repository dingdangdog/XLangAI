package repository

// VoiceRole 语音合成策略（与 manager schema synthesis_type 一致）。
const (
	SynthesisTTS              = "tts"
	SynthesisNativeAudioInText = "native_audio_in_text"
	SynthesisNativeAudioIO     = "native_audio_io"
)

func NormalizeSynthesisType(raw string) string {
	switch raw {
	case SynthesisNativeAudioInText, SynthesisNativeAudioIO:
		return raw
	default:
		return SynthesisTTS
	}
}

func IsNativeSynthesis(st string) bool {
	return st == SynthesisNativeAudioInText || st == SynthesisNativeAudioIO
}

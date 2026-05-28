package voiceaudio

import "errors"

var (
	ErrConfig           = errors.New("voice synth: invalid voice role config")
	ErrProviderNotReady = errors.New("voice synth: provider not ready")
	ErrUnsupported      = errors.New("voice synth: unsupported protocol")
	ErrFailed           = errors.New("voice synth: generation failed")
	ErrNoTranscript     = errors.New("voice synth: no transcript in audio response")
	ErrNoAudio          = errors.New("voice synth: no audio in response")
	ErrNoText           = errors.New("voice synth: no text in response")
)

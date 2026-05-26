package ai

import "testing"

func Test_ttsLoudnessOutputArgs(t *testing.T) {
	tests := []struct {
		mime    string
		format  string
		outMime string
	}{
		{"audio/mpeg", "mp3", "audio/mpeg"},
		{"audio/mp3", "mp3", "audio/mpeg"},
		{"audio/wav", "wav", "audio/wav"},
		{"audio/ogg", "ogg", "audio/ogg"},
		{"", "mp3", "audio/mpeg"},
	}
	for _, tc := range tests {
		format, _, outMime := ttsLoudnessOutputArgs(tc.mime)
		if format != tc.format || outMime != tc.outMime {
			t.Errorf("mime %q: got format=%s outMime=%s, want format=%s outMime=%s",
				tc.mime, format, outMime, tc.format, tc.outMime)
		}
	}
}

func TestTTSLoudnessOptions_resolved_defaults(t *testing.T) {
	i, tp, lra := (*TTSLoudnessOptions)(nil).resolved()
	if i != -14.0 || tp != -1.0 || lra != 8.0 {
		t.Fatalf("defaults: I=%g TP=%g LRA=%g", i, tp, lra)
	}
}

func TestTTSLoudnessOptions_resolved_override(t *testing.T) {
	o := &TTSLoudnessOptions{TargetLUFS: -14, TruePeakDB: -2, LRA: 9}
	i, tp, lra := o.resolved()
	if i != -14 || tp != -2 || lra != 9 {
		t.Fatalf("override: I=%g TP=%g LRA=%g", i, tp, lra)
	}
}

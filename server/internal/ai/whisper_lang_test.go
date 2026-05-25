package ai

import "testing"

func TestWhisperLanguageHintFromSysLanguageCode(t *testing.T) {
	cases := []struct {
		in   string
		want *string
	}{
		{"", nil},
		{"  ", nil},
		{"yue", strPtr("zh")},
		{"YUE", strPtr("zh")},
		{"cmn", strPtr("zh")},
		{"zh-HK", strPtr("zh")},
		{"en", strPtr("en")},
		{"en-US", strPtr("en")},
		{"ja", strPtr("ja")},
		{"ja-JP", strPtr("ja")},
		{"xxx", nil},
	}
	for _, tc := range cases {
		got := WhisperLanguageHintFromSysLanguageCode(tc.in)
		if !ptrEq(got, tc.want) {
			t.Fatalf("WhisperLanguageHintFromSysLanguageCode(%q): got %v want %v", tc.in, strVal(got), strVal(tc.want))
		}
	}
}

func strPtr(s string) *string { return &s }

func strVal(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func ptrEq(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

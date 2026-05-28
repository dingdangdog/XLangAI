package voiceaudio

import (
	"encoding/json"
	"strings"
)

type roleExtraConfig struct {
	ResponseModalities []string `json:"response_modalities"`
	InputModalities    []string `json:"input_modalities"`
}

func parseRoleConfig(raw *string) roleExtraConfig {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return roleExtraConfig{}
	}
	var c roleExtraConfig
	_ = json.Unmarshal([]byte(*raw), &c)
	return c
}

func geminiVoiceName(voiceCode string) string {
	v := strings.TrimSpace(voiceCode)
	if v == "" || v == "-" {
		return "Kore"
	}
	return v
}

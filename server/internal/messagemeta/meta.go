package messagemeta

import (
	"encoding/json"
	"strings"
)

// AI 交流状态（存于 usr_messages.metadata JSON）。
const (
	StatusSuccess        = "success"
	StatusQuotaExceeded  = "quota_exceeded"
	StatusFailed         = "failed"
)

type payload struct {
	AIInteractionStatus string `json:"ai_interaction_status"`
}

// MarshalStatus 写入 metadata JSON。
func MarshalStatus(status string) *string {
	if status == "" {
		return nil
	}
	b, err := json.Marshal(payload{AIInteractionStatus: status})
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

// ParseStatus 从 metadata 解析 AI 交流状态。
func ParseStatus(meta *string) string {
	if meta == nil {
		return ""
	}
	raw := strings.TrimSpace(*meta)
	if raw == "" {
		return ""
	}
	var p payload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return ""
	}
	return p.AIInteractionStatus
}

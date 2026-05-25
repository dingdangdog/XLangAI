package oauth

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/idtoken"
)

// VerifyGoogleIDToken 校验 Google id_token，返回 sub 与可选 email。
// audiences 为允许的 OAuth 2.0 Client ID 列表（Android / iOS / Web 等），任一匹配即通过。
func VerifyGoogleIDToken(ctx context.Context, raw string, audiences []string) (sub string, email *string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, errors.New("empty id_token")
	}
	var lastErr error
	for _, aud := range audiences {
		aud = strings.TrimSpace(aud)
		if aud == "" {
			continue
		}
		payload, err := idtoken.Validate(ctx, raw, aud)
		if err != nil {
			lastErr = err
			continue
		}
		s := payload.Subject
		if s == "" {
			return "", nil, errors.New("missing sub")
		}
		var em *string
		if e, ok := payload.Claims["email"].(string); ok && strings.TrimSpace(e) != "" {
			v := strings.TrimSpace(e)
			em = &v
		}
		return s, em, nil
	}
	if lastErr != nil {
		return "", nil, lastErr
	}
	return "", nil, errors.New("no audience configured")
}

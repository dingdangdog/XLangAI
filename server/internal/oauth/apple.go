package oauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const appleJWKSURL = "https://appleid.apple.com/auth/keys"
const appleIssuer = "https://appleid.apple.com"

type appleJWKSet struct {
	Keys []struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		Crv string `json:"crv"`
		X   string `json:"x"`
		Y   string `json:"y"`
	} `json:"keys"`
}

var (
	appleJWKSMu   sync.RWMutex
	appleJWKSKeys map[string]*ecdsa.PublicKey
	appleJWKSFetched time.Time
)

func b64url(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	// RawURLEncoding 无 padding；Apple 使用无 padding
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}

func jwkECP256Key(kid, crv, xb, yb string) (*ecdsa.PublicKey, error) {
	if strings.TrimSpace(kid) == "" {
		return nil, errors.New("empty kid")
	}
	if crv != "P-256" {
		return nil, fmt.Errorf("unsupported crv %q", crv)
	}
	xRaw, err := b64url(xb)
	if err != nil {
		return nil, err
	}
	yRaw, err := b64url(yb)
	if err != nil {
		return nil, err
	}
	x := new(big.Int).SetBytes(xRaw)
	y := new(big.Int).SetBytes(yRaw)
	if !elliptic.P256().IsOnCurve(x, y) {
		return nil, errors.New("point not on P-256")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

func refreshAppleJWKS(ctx context.Context) (map[string]*ecdsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appleJWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("apple jwks: status %d: %s", resp.StatusCode, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var set appleJWKSet
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, err
	}
	out := make(map[string]*ecdsa.PublicKey)
	for _, k := range set.Keys {
		if k.Kty != "EC" {
			continue
		}
		pub, err := jwkECP256Key(k.Kid, k.Crv, k.X, k.Y)
		if err != nil {
			continue
		}
		out[k.Kid] = pub
	}
	if len(out) == 0 {
		return nil, errors.New("no usable EC keys in Apple JWKS")
	}
	return out, nil
}

func applePublicKeyForKid(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	appleJWKSMu.RLock()
	if pub, ok := appleJWKSKeys[kid]; ok && time.Since(appleJWKSFetched) < 6*time.Hour {
		appleJWKSMu.RUnlock()
		return pub, nil
	}
	appleJWKSMu.RUnlock()

	appleJWKSMu.Lock()
	defer appleJWKSMu.Unlock()
	if pub, ok := appleJWKSKeys[kid]; ok && time.Since(appleJWKSFetched) < 6*time.Hour {
		return pub, nil
	}
	m, err := refreshAppleJWKS(ctx)
	if err != nil {
		return nil, err
	}
	appleJWKSKeys = m
	appleJWKSFetched = time.Now()
	pub, ok := m[kid]
	if !ok {
		return nil, fmt.Errorf("unknown apple jwk kid %q", kid)
	}
	return pub, nil
}

// VerifyAppleIDToken 校验 Apple identityToken（JWT），返回 sub 与可选 email。
func VerifyAppleIDToken(ctx context.Context, raw string, audiences []string) (sub string, email *string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil, errors.New("empty identity_token")
	}
	if len(audiences) == 0 {
		return "", nil, errors.New("no audience configured")
	}
	token, _, err := jwt.NewParser().ParseUnverified(raw, jwt.MapClaims{})
	if err != nil {
		return "", nil, err
	}
	kid, _ := token.Header["kid"].(string)
	if strings.TrimSpace(kid) == "" {
		return "", nil, errors.New("missing kid in token header")
	}
	pub, err := applePublicKeyForKid(ctx, kid)
	if err != nil {
		return "", nil, err
	}
	parsed, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return pub, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
		jwt.WithIssuer(appleIssuer),
		jwt.WithLeeway(2*time.Minute),
	)
	if err != nil {
		return "", nil, err
	}
	if !parsed.Valid {
		return "", nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil, errors.New("invalid claims")
	}
	if !audienceAllowed(claims["aud"], audiences) {
		return "", nil, errors.New("audience mismatch")
	}
	sv, _ := claims["sub"].(string)
	if strings.TrimSpace(sv) == "" {
		return "", nil, errors.New("missing sub")
	}
	if e, ok := claims["email"].(string); ok && strings.TrimSpace(e) != "" {
		v := strings.TrimSpace(e)
		email = &v
	}
	return sv, email, nil
}

func audienceAllowed(raw any, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a != "" {
			set[a] = struct{}{}
		}
	}
	if len(set) == 0 {
		return false
	}
	switch v := raw.(type) {
	case string:
		_, ok := set[strings.TrimSpace(v)]
		return ok
	case []any:
		for _, x := range v {
			s, _ := x.(string)
			if _, ok := set[strings.TrimSpace(s)]; ok {
				return true
			}
		}
		return false
	default:
		return false
	}
}

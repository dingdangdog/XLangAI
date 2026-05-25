package billing

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	appleProdHost    = "https://api.storekit.itunes.apple.com"
	appleSandboxHost = "https://api.storekit-sandbox.itunes.apple.com"
)

// AppleConfig App Store Server API（交易查询）配置。
type AppleConfig struct {
	IssuerID   string
	KeyID      string
	BundleID   string
	PrivateKey *ecdsa.PrivateKey
	Sandbox    bool
}

func LoadApplePrivateKeyFromFile(path string) (*ecdsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseApplePrivateKeyPEM(b)
}

// ParseApplePrivateKeyPEM 解析 App Store Connect 下载的 AuthKey_xxx.p8（PKCS#8 EC）。
func ParseApplePrivateKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("apple private key: invalid PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("apple private key: expected ECDSA key")
	}
	return ec, nil
}

func (c *AppleConfig) signedToken() (string, error) {
	if c == nil || c.PrivateKey == nil || strings.TrimSpace(c.IssuerID) == "" || strings.TrimSpace(c.KeyID) == "" || strings.TrimSpace(c.BundleID) == "" {
		return "", errors.New("apple billing: incomplete configuration")
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss": c.IssuerID,
		"iat": now.Unix(),
		"exp": now.Add(19 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": c.BundleID,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	t.Header["kid"] = c.KeyID
	return t.SignedString(c.PrivateKey)
}

// IOSignedTransactionPayload 解码自 Apple 返回的 signedTransactionInfo JWS payload（未单独验签；依赖 TLS 与本服务持有的 App Store Connect API 密钥）。
type IOSignedTransactionPayload struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	ProductID             string `json:"productId"`
	BundleID              string `json:"bundleId"`
	ExpiresDate           int64  `json:"expiresDate"` // ms since epoch, 0 for consumable
}

func decodeJWSPayload(jws string, out any) error {
	parts := strings.Split(jws, ".")
	if len(parts) < 2 {
		return errors.New("invalid jws")
	}
	payload := parts[1]
	payload = strings.ReplaceAll(payload, "-", "+")
	payload = strings.ReplaceAll(payload, "_", "/")
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

type appleTxnAPIResponse struct {
	SignedTransactionInfo string `json:"signedTransactionInfo"`
}

// FetchIOSTransaction 使用 transactionId 查询 Apple，返回解码后的交易载荷。
func FetchIOSTransaction(ctx context.Context, cfg *AppleConfig, transactionID string) (*IOSignedTransactionPayload, error) {
	if strings.TrimSpace(transactionID) == "" {
		return nil, errors.New("missing transaction id")
	}
	bearer, err := cfg.signedToken()
	if err != nil {
		return nil, err
	}
	tryHosts := []string{appleProdHost, appleSandboxHost}
	if cfg.Sandbox {
		tryHosts = []string{appleSandboxHost, appleProdHost}
	}
	var lastErr error
	for _, host := range tryHosts {
		url := fmt.Sprintf("%s/inApps/v1/transactions/%s", strings.TrimRight(host, "/"), transactionID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+bearer)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized {
			lastErr = fmt.Errorf("apple http %d: %s", resp.StatusCode, string(body))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("apple http %d: %s", resp.StatusCode, string(body))
		}
		var wrap appleTxnAPIResponse
		if err := json.Unmarshal(body, &wrap); err != nil {
			return nil, err
		}
		if strings.TrimSpace(wrap.SignedTransactionInfo) == "" {
			return nil, errors.New("apple: empty signedTransactionInfo")
		}
		var payload IOSignedTransactionPayload
		if err := decodeJWSPayload(wrap.SignedTransactionInfo, &payload); err != nil {
			return nil, err
		}
		return &payload, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("apple: transaction not found")
}

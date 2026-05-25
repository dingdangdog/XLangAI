package ai

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TencentCloudTC3 调用腾讯云 API 3.0（TC3-HMAC-SHA256）。
func TencentCloudTC3(
	ctx context.Context,
	secretID, secretKey, region, endpoint, service, action, version string,
	payload map[string]any,
) ([]byte, error) {
	secretID = strings.TrimSpace(secretID)
	secretKey = strings.TrimSpace(secretKey)
	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("tencent cloud: missing secret id or key")
	}
	if region == "" {
		region = "ap-guangzhou"
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = service + ".tencentcloudapi.com"
	}
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	endpoint = strings.TrimRight(endpoint, "/")

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	canonicalHeaders := "content-type:application/json; charset=utf-8\nhost:" + endpoint + "\n"
	signedHeaders := "content-type;host"
	hashedPayload := sha256HexBytes(body)
	canonicalRequest := "POST\n/\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + hashedPayload
	credentialScope := date + "/" + service + "/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + fmt.Sprintf("%d", timestamp) + "\n" + credentialScope + "\n" + sha256HexBytes([]byte(canonicalRequest))
	secretDate := hmacSHA256Bytes([]byte("TC3"+secretKey), []byte(date))
	secretService := hmacSHA256Bytes(secretDate, []byte(service))
	secretSigning := hmacSHA256Bytes(secretService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSHA256Bytes(secretSigning, []byte(stringToSign)))
	auth := fmt.Sprintf(
		"TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		secretID, credentialScope, signedHeaders, signature,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", endpoint)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Region", region)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tencent %s %s: %s: %s", service, action, resp.Status, truncateTTSErr(b))
	}
	return b, nil
}

func sha256HexBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hmacSHA256Bytes(key, msg []byte) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write(msg)
	return m.Sum(nil)
}

package translate

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
	"strconv"
	"strings"
	"time"
)

func translateTencent(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	secretID := strings.TrimSpace(in.APIKey)
	secretKey := strings.TrimSpace(in.APISecret)
	legacyID := strings.TrimSpace(in.Config.AppID)
	if secretKey == "" && legacyID != "" {
		// 旧配置：SecretId → config.app_id，SecretKey → api_key
		secretID = legacyID
		secretKey = strings.TrimSpace(in.APIKey)
	}
	if secretID == "" || secretKey == "" {
		return "", ErrProviderNotReady
	}
	region := strings.TrimSpace(in.Config.Region)
	if region == "" {
		region = "ap-guangzhou"
	}
	endpoint := "tmt.tencentcloudapi.com"
	if bu := strings.TrimSpace(in.BaseURL); bu != "" {
		bu = strings.TrimPrefix(strings.TrimPrefix(bu, "https://"), "http://")
		bu = strings.TrimRight(bu, "/")
		if bu != "" {
			endpoint = bu
		}
	}
	to := TargetForProvider("tencent_translate", targetLocale)
	projectID := int64(0)
	if pid := strings.TrimSpace(in.Config.ProjectID); pid != "" {
		if v, err := strconv.ParseInt(pid, 10, 64); err == nil {
			projectID = v
		}
	}
	payload := map[string]any{
		"SourceText": text,
		"Source":     "auto",
		"Target":     to,
		"ProjectId":  projectID,
	}
	body, _ := json.Marshal(payload)
	service := "tmt"
	action := "TextTranslate"
	version := "2018-03-21"
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	canonicalHeaders := "content-type:application/json; charset=utf-8\nhost:" + endpoint + "\n"
	signedHeaders := "content-type;host"
	hashedPayload := sha256Hex(string(body))
	canonicalRequest := "POST\n/\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + hashedPayload
	credentialScope := date + "/" + service + "/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + fmt.Sprintf("%d", timestamp) + "\n" + credentialScope + "\n" + sha256Hex(canonicalRequest)
	secretDate := hmacSHA256([]byte("TC3"+secretKey), []byte(date))
	secretService := hmacSHA256(secretDate, []byte(service))
	secretSigning := hmacSHA256(secretService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSHA256(secretSigning, []byte(stringToSign)))
	auth := fmt.Sprintf(
		"TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		secretID, credentialScope, signedHeaders, signature,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
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
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tencent translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		Response struct {
			TargetText string `json:"TargetText"`
			Error      *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.Response.Error != nil {
		return "", fmt.Errorf("tencent translate: %s", parsed.Response.Error.Message)
	}
	return strings.TrimSpace(parsed.Response.TargetText), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, msg []byte) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write(msg)
	return m.Sum(nil)
}

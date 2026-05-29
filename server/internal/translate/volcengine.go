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
	"net/url"
	"sort"
	"strings"
	"time"
)

func translateVolcengine(ctx context.Context, in ServiceInput, text, targetLocale string) (string, error) {
	accessKeyID := strings.TrimSpace(in.APIKey)
	secretAccessKey := strings.TrimSpace(in.APISecret)
	if accessKeyID == "" || secretAccessKey == "" {
		return "", ErrProviderNotReady
	}
	region := strings.TrimSpace(in.Config.Region)
	if region == "" {
		region = "cn-north-1"
	}
	host := "open.volcengineapi.com"
	endpoint := "https://" + host
	if bu := strings.TrimSpace(in.BaseURL); bu != "" && strings.HasPrefix(strings.ToLower(bu), "http") {
		endpoint = strings.TrimRight(bu, "/")
	}
	to := TargetForProvider("volcengine_translate", targetLocale)
	body, _ := json.Marshal(map[string]any{
		"TargetLanguage": to,
		"TextList":       []string{text},
	})
	query := url.Values{}
	query.Set("Action", "TranslateText")
	query.Set("Version", "2020-06-01")
	now := time.Now().UTC()
	date := now.Format("20060102T150405Z")
	auth, signedQuery := volcengineSign(accessKeyID, secretAccessKey, region, "POST", "/", query, string(body), date)
	reqURL := endpoint + "/?" + signedQuery
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Date", date)
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("volcengine translate: %s: %s", resp.Status, truncateErr(b, 512))
	}
	var parsed struct {
		TranslationList []struct {
			Translation string `json:"Translation"`
		} `json:"TranslationList"`
		ResponseMetadata struct {
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"ResponseMetadata"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", err
	}
	if parsed.ResponseMetadata.Error.Code != "" {
		return "", fmt.Errorf("volcengine translate: %s %s", parsed.ResponseMetadata.Error.Code, parsed.ResponseMetadata.Error.Message)
	}
	if len(parsed.TranslationList) == 0 {
		return "", fmt.Errorf("volcengine translate: empty result")
	}
	return strings.TrimSpace(parsed.TranslationList[0].Translation), nil
}

func volcengineSign(accessKeyID, secretKey, region, method, path string, query url.Values, body, date string) (string, string) {
	service := "translate"
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonicalQuery strings.Builder
	for i, k := range keys {
		if i > 0 {
			canonicalQuery.WriteByte('&')
		}
		canonicalQuery.WriteString(volcPercentEncode(k))
		canonicalQuery.WriteByte('=')
		canonicalQuery.WriteString(volcPercentEncode(query.Get(k)))
	}
	qStr := canonicalQuery.String()
	payloadHash := sha256Hex(body)
	canonicalRequest := method + "\n" + path + "\n" + qStr + "\n" +
		"content-type:application/json\n" +
		"host:open.volcengineapi.com\n" +
		"x-date:" + date + "\n\n" +
		"content-type;host;x-date\n" + payloadHash
	credentialScope := date[:8] + "/" + region + "/" + service + "/request"
	stringToSign := "HMAC-SHA256\n" + date + "\n" + credentialScope + "\n" + sha256Hex(canonicalRequest)
	kDate := volcHMAC([]byte(secretKey), []byte(date[:8]))
	kRegion := volcHMAC(kDate, []byte(region))
	kService := volcHMAC(kRegion, []byte(service))
	kSigning := volcHMAC(kService, []byte("request"))
	signature := hex.EncodeToString(volcHMAC(kSigning, []byte(stringToSign)))
	signedQuery := qStr + "&X-Date=" + volcPercentEncode(date) + "&X-NotSignBody=&X-Credential=" + volcPercentEncode(accessKeyID+"/"+credentialScope) +
		"&X-Algorithm=HMAC-SHA256&X-SignedHeaders=content-type;host;x-date&X-SignedQueries=Action;Version&X-Signature=" + volcPercentEncode(signature)
	auth := "HMAC-SHA256 Credential=" + accessKeyID + "/" + credentialScope + ", SignedHeaders=content-type;host;x-date, Signature=" + signature
	return auth, signedQuery
}

func volcHMAC(key, msg []byte) []byte {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write(msg)
	return m.Sum(nil)
}

func volcPercentEncode(s string) string {
	var buf strings.Builder
	for _, b := range []byte(s) {
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' || b == '~' {
			buf.WriteByte(b)
		} else {
			buf.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}
	return buf.String()
}

package translate

import (
	"encoding/hex"
	"time"
)

func aws4SigningKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func aws4Authorization(accessKeyID, secretAccessKey, region, service, host, method, canonicalURI, payload string, extraHeaders map[string]string) (string, string, error) {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	var headerLines []string
	headerLines = append(headerLines, "content-type:application/x-amz-json-1.1")
	headerLines = append(headerLines, "host:"+host)
	for k, v := range extraHeaders {
		headerLines = append(headerLines, k+":"+v)
	}
	headerLines = append(headerLines, "x-amz-date:"+amzDate)

	signedHeaderKeys := []string{"content-type", "host"}
	for k := range extraHeaders {
		signedHeaderKeys = append(signedHeaderKeys, k)
	}
	signedHeaderKeys = append(signedHeaderKeys, "x-amz-date")
	// stable order for signing
	sortStrings(signedHeaderKeys)

	canonicalHeaders := ""
	for _, k := range signedHeaderKeys {
		switch k {
		case "content-type":
			canonicalHeaders += "content-type:application/x-amz-json-1.1\n"
		case "host":
			canonicalHeaders += "host:" + host + "\n"
		case "x-amz-date":
			canonicalHeaders += "x-amz-date:" + amzDate + "\n"
		default:
			if v, ok := extraHeaders[k]; ok {
				canonicalHeaders += k + ":" + v + "\n"
			}
		}
	}
	signedHeaders := joinSemicolon(signedHeaderKeys)

	payloadHash := sha256Hex(payload)
	canonicalRequest := method + "\n" + canonicalURI + "\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHash
	credentialScope := dateStamp + "/" + region + "/" + service + "/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256Hex(canonicalRequest)
	signingKey := aws4SigningKey(secretAccessKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))
	auth := "AWS4-HMAC-SHA256 Credential=" + accessKeyID + "/" + credentialScope + ", SignedHeaders=" + signedHeaders + ", Signature=" + signature
	return auth, amzDate, nil
}

func sortStrings(ss []string) {
	for i := 0; i < len(ss); i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[j] < ss[i] {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
}

func joinSemicolon(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for i := 1; i < len(ss); i++ {
		out += ";" + ss[i]
	}
	return out
}

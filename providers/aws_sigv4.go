package providers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

// SignPollyRequest signs req with AWS Signature Version 4 for the Polly service.
// accessKeyID and secretAccessKey are the AWS credentials.
// region is the AWS region (e.g. us-east-1).
func SignPollyRequest(req *http.Request, body []byte, accessKeyID, secretAccessKey, region string) error {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	payloadHash := sha256Hash(body)
	payloadHashHex := hex.EncodeToString(payloadHash)

	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	req.Header.Set("Host", host)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHashHex)

	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQuery := ""
	canonicalHeaders := "content-type:application/json\n"
	canonicalHeaders += "host:" + host + "\n"
	canonicalHeaders += "x-amz-content-sha256:" + payloadHashHex + "\n"
	canonicalHeaders += "x-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHashHex,
	}, "\n")

	credentialScope := dateStamp + "/" + region + "/polly/aws4_request"
	canonicalRequestHash := sha256Hash([]byte(canonicalRequest))
	canonicalRequestHashHex := hex.EncodeToString(canonicalRequestHash)

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		canonicalRequestHashHex,
	}, "\n")

	signingKey := deriveSigningKey(secretAccessKey, dateStamp, region, "polly")
	signature := hmacSHA256Hex(signingKey, []byte(stringToSign))

	authHeader := "AWS4-HMAC-SHA256 Credential=" + accessKeyID + "/" + credentialScope +
		", SignedHeaders=" + signedHeaders +
		", Signature=" + signature
	req.Header.Set("Authorization", authHeader)

	return nil
}

func sha256Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func hmacSHA256Hex(key []byte, data []byte) string {
	return hex.EncodeToString(hmacSHA256(key, data))
}

func deriveSigningKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}
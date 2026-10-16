package dns

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func (p *Route53Provider) sign(request *http.Request, payloadHash string) {
	moment := p.now().UTC()
	amzDate := moment.Format("20060102T150405Z")
	dateStamp := moment.Format("20060102")
	request.Header.Set("X-Amz-Date", amzDate)
	headers := []string{"host", "x-amz-date"}
	if p.session != "" {
		request.Header.Set("X-Amz-Security-Token", p.session)
		headers = append(headers, "x-amz-security-token")
	}
	var canonicalHeaders strings.Builder
	values := map[string]string{
		"host":       request.URL.Host,
		"x-amz-date": amzDate,
	}
	if p.session != "" {
		values["x-amz-security-token"] = p.session
	}
	for _, name := range headers {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(values[name]))
		canonicalHeaders.WriteString("\n")
	}
	query := request.URL.Query().Encode()
	canonical := strings.Join([]string{request.Method, request.URL.EscapedPath(), query, canonicalHeaders.String(), strings.Join(headers, ";"), payloadHash}, "\n")
	scope := strings.Join([]string{dateStamp, "us-east-1", "route53", "aws4_request"}, "/")
	toSign := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, hex.EncodeToString(hashSHA256(canonical))}, "\n")
	signingKey := hmacSHA256(hmacSHA256(hmacSHA256(hmacSHA256([]byte("AWS4"+p.secretKey), dateStamp), "us-east-1"), "route53"), "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(signingKey, toSign))
	request.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s", p.accessKey, scope, strings.Join(headers, ";"), signature))
}

func hashSHA256(data string) []byte {
	sum := sha256.Sum256([]byte(data))
	return sum[:]
}

type route53Error struct {
	XMLName xml.Name `xml:"ErrorResponse"`
	Code    string   `xml:"Error>Code"`
	Message string   `xml:"Error>Message"`
}

func (p *Route53Provider) call(ctx context.Context, method, path, query string, body []byte) ([]byte, error) {
	target := p.baseURL + path
	if query != "" {
		target += "?" + query
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, fmt.Errorf("create route53 request: %w", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/xml")
	}
	payloadHash := hex.EncodeToString(hashSHA256(string(body)))
	p.sign(request, payloadHash)
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("route53 request failed: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read route53 response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiErr route53Error
		if err := xml.Unmarshal(raw, &apiErr); err == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("route53 %s: %s", apiErr.Code, apiErr.Message)
		}
		return nil, fmt.Errorf("route53 returned HTTP %d", response.StatusCode)
	}
	return raw, nil
}

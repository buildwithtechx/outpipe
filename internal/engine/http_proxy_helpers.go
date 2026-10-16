package engine

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"outpipe.dev/outpipe/internal/security"
)

func authorizedRequest(request *http.Request, passwordHash string) bool {
	_, password, ok := request.BasicAuth()
	return ok && security.VerifyPassword(password, passwordHash)
}

func clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(request.RemoteAddr)
}

var errBodyTooLarge = fmt.Errorf("request body too large")

func readBody(body io.ReadCloser, maxBytes int64) (string, error) {
	if body == nil {
		return "", nil
	}
	defer body.Close()
	reader := io.Reader(body)
	if maxBytes > 0 {
		reader = io.LimitReader(body, maxBytes+1)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read request body: %w", err)
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return "", errBodyTooLarge
	}
	if len(data) == 0 {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func resolveTunnelID(host string, baseDomain string) (string, bool) {
	cleanHost := strings.ToLower(strings.TrimSuffix(host, "."))
	if parsedHost, _, err := net.SplitHostPort(cleanHost); err == nil {
		cleanHost = parsedHost
	}
	baseDomain = strings.ToLower(strings.TrimSuffix(baseDomain, "."))
	prefix, found := strings.CutSuffix(cleanHost, "."+baseDomain)
	if !found || prefix == "" || strings.Contains(prefix, ".") || prefix == "www" {
		return "", false
	}
	return prefix, true
}

func resolveRouteKey(host string, baseDomain string) (string, bool) {
	cleanHost := strings.ToLower(strings.TrimSuffix(host, "."))
	if parsedHost, _, err := net.SplitHostPort(cleanHost); err == nil {
		cleanHost = parsedHost
	}
	baseDomain = strings.ToLower(strings.TrimSuffix(baseDomain, "."))
	prefix, found := strings.CutSuffix(cleanHost, "."+baseDomain)
	if found {
		if prefix == "" || strings.Contains(prefix, ".") || prefix == "www" {
			return "", false
		}
		return prefix, true
	}
	if cleanHost == "" || strings.Contains(cleanHost, " ") || strings.Contains(cleanHost, "..") || cleanHost == "www" {
		return "", false
	}
	return cleanHost, true
}

func requestHeaders(headers http.Header) map[string][]string {
	result := make(map[string][]string)
	for key, values := range headers {
		if isHopByHopHeader(key) {
			continue
		}
		result[key] = append([]string(nil), values...)
	}
	return result
}

func writeResponseHeaders(destination http.Header, source map[string][]string) {
	for key, values := range source {
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

func isHopByHopHeader(key string) bool {
	switch strings.ToLower(key) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

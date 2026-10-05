package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

const MaxCaptureBytes = 64 * 1024

type RequestCapture struct {
	OrganizationID   string    `json:"organization_id"`
	TunnelID         string    `json:"tunnel_id"`
	Timestamp        time.Time `json:"timestamp"`
	Method           string    `json:"method"`
	Path             string    `json:"path"`
	StatusCode       int       `json:"status_code"`
	DurationMs       int64     `json:"duration_ms"`
	RequestHeaders   string    `json:"request_headers"`
	ResponseHeaders  string    `json:"response_headers"`
	RequestBody      string    `json:"request_body"`
	ResponseBody     string    `json:"response_body"`
	RequestBodySize  int64     `json:"request_body_size"`
	ResponseBodySize int64     `json:"response_body_size"`
}

type CaptureRecorder interface {
	Enabled(context.Context, string, string) (bool, error)
	RecordCapture(context.Context, RequestCapture) error
}

func (p *HTTPProxy) SetCaptureRecorder(recorder CaptureRecorder) { p.capture = recorder }

func (p *HTTPProxy) refreshCapture(ctx context.Context, route string) {
	if p.capture == nil {
		return
	}
	tunnelID, ok := p.router.sessions.Resolve(route)
	if !ok {
		return
	}
	orgID, ok := p.router.OrganizationID(route)
	if !ok {
		return
	}
	enabled, err := p.capture.Enabled(ctx, tunnelID, orgID)
	p.router.SetCaptureEnabled(route, err == nil && enabled)
}

func (p *HTTPProxy) captureRequest(request *http.Request, route, encodedBody string, response protocol.HTTPResponse, started time.Time) error {
	if p.capture == nil || !p.router.IsCaptureEnabled(route) {
		return nil
	}
	tunnelID, ok := p.router.sessions.Resolve(route)
	if !ok {
		return nil
	}
	orgID, ok := p.router.OrganizationID(route)
	if !ok {
		return nil
	}
	reqHeaders, err := json.Marshal(redactCaptureHeaders(request.Header))
	if err != nil {
		return fmt.Errorf("encode capture request headers: %w", err)
	}
	respHeaders, err := json.Marshal(redactCaptureHeaders(response.Headers))
	if err != nil {
		return fmt.Errorf("encode capture response headers: %w", err)
	}
	if len(reqHeaders) > MaxCaptureBytes {
		reqHeaders = []byte("{}")
	}
	if len(respHeaders) > MaxCaptureBytes {
		respHeaders = []byte("{}")
	}
	body, err := base64.StdEncoding.DecodeString(encodedBody)
	if err != nil {
		return fmt.Errorf("decode capture request: %w", err)
	}
	respBody, err := base64.StdEncoding.DecodeString(response.Body)
	if err != nil {
		return fmt.Errorf("decode capture response: %w", err)
	}
	capture := RequestCapture{OrganizationID: orgID, TunnelID: tunnelID, Timestamp: started.UTC(), Method: request.Method, Path: request.URL.Path, StatusCode: response.StatusCode, DurationMs: time.Since(started).Milliseconds(), RequestHeaders: string(reqHeaders), ResponseHeaders: string(respHeaders), RequestBody: sanitizedCaptureBody(body), ResponseBody: sanitizedCaptureBody(respBody), RequestBodySize: int64(len(body)), ResponseBodySize: int64(len(respBody))}
	if err := p.capture.RecordCapture(request.Context(), capture); err != nil {
		return fmt.Errorf("record request capture: %w", err)
	}
	return nil
}

func sensitiveCaptureKey(key string) bool {
	key = strings.ToLower(key)
	for _, value := range []string{"authorization", "cookie", "token", "password", "secret", "credential", "api-key", "api_key", "apikey"} {
		if strings.Contains(key, value) {
			return true
		}
	}
	return false
}

func redactCaptureHeaders(headers map[string][]string) map[string][]string {
	result := make(map[string][]string)
	size := 0
	for key, values := range headers {
		if sensitiveCaptureKey(key) || isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			size += len(key) + len(value)
			if size > MaxCaptureBytes {
				return result
			}
			result[key] = append(result[key], value)
		}
	}
	return result
}

func sanitizedCaptureBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if len(body) > MaxCaptureBytes {
		return "[payload exceeds capture limit]"
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return "[non-JSON payload omitted]"
	}
	redactCaptureValue(value, 0)
	data, err := json.Marshal(value)
	if err != nil {
		return "[payload omitted]"
	}
	if len(data) > MaxCaptureBytes {
		return "[payload exceeds capture limit]"
	}
	return string(data)
}

func redactCaptureValue(value any, depth int) {
	if depth > 32 {
		return
	}
	switch value := value.(type) {
	case map[string]any:
		for key, item := range value {
			if sensitiveCaptureKey(key) || depth == 32 {
				value[key] = "[redacted]"
			} else {
				redactCaptureValue(item, depth+1)
			}
		}
	case []any:
		for i, item := range value {
			if depth == 32 {
				value[i] = "[redacted]"
			} else {
				redactCaptureValue(item, depth+1)
			}
		}
	}
}

package engine

import (
	"strings"
	"testing"
)

func TestCaptureRedactsNestedCredentialsAndBoundsBodies(t *testing.T) {
	body := sanitizedCaptureBody([]byte(`{"nested":[{"password":"hidden","apiKey":"hidden","value":42}],"token":"hidden"}`))
	if strings.Contains(body, "hidden") || !strings.Contains(body, "42") {
		t.Fatal("credential redaction lost safe data or leaked secrets")
	}
	if body := sanitizedCaptureBody([]byte("password=hidden")); strings.Contains(body, "hidden") {
		t.Fatal("non-JSON credentials captured")
	}
	if body := sanitizedCaptureBody([]byte(strings.Repeat("x", MaxCaptureBytes+1))); len(body) > MaxCaptureBytes {
		t.Fatal("capture exceeds limit")
	}
	headers := redactCaptureHeaders(map[string][]string{"Authorization": {"Bearer hidden"}, "Cookie": {"hidden"}, "X-Api-Key": {"hidden"}, "Content-Type": {"application/json"}})
	if len(headers) != 1 || headers["Content-Type"][0] != "application/json" {
		t.Fatal("capture headers leaked credential fields")
	}
}

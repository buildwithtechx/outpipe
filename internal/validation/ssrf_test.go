package validation

import (
	"net/http"
	"testing"
)

func TestValidateWebhookURLRejectsUnsafeLiteralHosts(t *testing.T) {
	for _, rawURL := range []string{
		"http://127.0.0.1:8080/hook",
		"http://[::1]:8080/hook",
		"http://169.254.169.254/latest/meta-data",
		"http://10.0.0.5/hook",
		"http://user:password@example.com/hook",
	} {
		if err := ValidateWebhookURL(rawURL); err == nil {
			t.Errorf("ValidateWebhookURL(%q) succeeded; want an error", rawURL)
		}
	}
}

func TestValidateWebhookURLAcceptsPublicSyntax(t *testing.T) {
	for _, rawURL := range []string{
		"https://hooks.example.com/outpipe",
		"http://hooks.example.com:8080/events?source=outpipe",
	} {
		if err := ValidateWebhookURL(rawURL); err != nil {
			t.Errorf("ValidateWebhookURL(%q) returned error: %v", rawURL, err)
		}
	}
}

func TestValidateURLSyntaxRejectsMalformedURLs(t *testing.T) {
	for _, rawURL := range []string{"", "example.com/hook", "ftp://example.com/hook", "http://"} {
		if err := ValidateURLSyntax(rawURL); err == nil {
			t.Errorf("ValidateURLSyntax(%q) succeeded; want an error", rawURL)
		}
	}
}

func TestNewSafeHTTPClientDoesNotFollowRedirects(t *testing.T) {
	client := NewSafeHTTPClient(0)

	if err := client.CheckRedirect(&http.Request{}, nil); err != http.ErrUseLastResponse {
		t.Fatalf("CheckRedirect() error = %v, want %v", err, http.ErrUseLastResponse)
	}

	transport, ok := client.Transport.(*http.Transport)

	if !ok {
		t.Fatalf("client transport type = %T, want *http.Transport", client.Transport)
	}

	if transport.Proxy != nil {
		t.Fatal("safe client must not use an ambient proxy")
	}
}

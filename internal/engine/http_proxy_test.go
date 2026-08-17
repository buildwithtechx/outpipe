package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/security"
	"outpipe.dev/outpipe/pkg/protocol"
)

func TestResolveTunnelID(t *testing.T) {
	tests := []struct {
		host     string
		expected string
		valid    bool
	}{
		{host: "abc.tunnel.localhost", expected: "abc", valid: true},
		{host: "abc.tunnel.localhost:443", expected: "abc", valid: true},
		{host: "tunnel.localhost", valid: false},
		{host: "abc.other.localhost", valid: false},
		{host: "www.tunnel.localhost", valid: false},
	}
	for _, test := range tests {
		actual, valid := resolveTunnelID(test.host, "tunnel.localhost")
		if valid != test.valid || actual != test.expected {
			t.Fatalf("resolveTunnelID(%q) = %q, %t", test.host, actual, valid)
		}
	}
}

func TestHTTPProxyRequiresTunnelPassword(t *testing.T) {
	hash, err := security.HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	sessions := NewSessionRegistry()
	if err := sessions.Reserve(Session{ID: "session", OrganizationID: "org", TunnelID: "protected", PasswordHash: hash, Send: func(context.Context, protocol.Envelope) error { return nil }}, false); err != nil {
		t.Fatalf("reserve session: %v", err)
	}
	router, err := NewRequestRouter(sessions, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	proxy, err := NewHTTPProxy("tunnel.localhost", router, 1024)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "http://protected.tunnel.localhost/", nil)
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got %d", response.Code)
	}
	request.SetBasicAuth("ignored", "secret-pass")
	response = httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	if response.Code == http.StatusUnauthorized {
		t.Fatal("expected valid password to pass the access check")
	}
}

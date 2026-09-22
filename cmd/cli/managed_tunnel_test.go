package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

func TestResolveManagedTunnel(t *testing.T) {
	cfg := config.CLIConfig{APIKey: "", APIURL: "http://api.test"}

	if _, err := resolveManagedTunnel(t.Context(), cfg, "t-1", ""); err == nil {
		t.Fatal("expected error without API key")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/tunnels/t-active" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"t-active","status":"active","publicHostname":"my-app.outpipe.app"}`))
			return
		}

		if r.URL.Path == "/api/v1/tunnels/t-revoked" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"t-revoked","status":"revoked","publicHostname":"my-app.outpipe.app"}`))
			return
		}

		if r.URL.Path == "/api/v1/tunnels/t-empty" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"t-empty","status":"active","publicHostname":""}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()
	cfg = config.CLIConfig{APIKey: "key-1", APIURL: server.URL}

	hostname, err := resolveManagedTunnel(t.Context(), cfg, "t-active", "")

	if err != nil {
		t.Fatal(err)
	}

	if hostname != "my-app.outpipe.app" {
		t.Errorf("expected hostname, got %q", hostname)
	}

	if _, err := resolveManagedTunnel(t.Context(), cfg, "t-revoked", ""); err == nil {
		t.Fatal("expected revoked tunnel to error")
	}

	if _, err := resolveManagedTunnel(t.Context(), cfg, "t-empty", ""); err == nil {
		t.Fatal("expected empty hostname to error")
	}

	if _, err := resolveManagedTunnel(t.Context(), cfg, "t-missing", ""); err == nil {
		t.Fatal("expected missing tunnel to error")
	}

	subdomain, err := resolveManagedTunnel(t.Context(), cfg, "t-active", "custom.example.com")

	if err != nil {
		t.Fatal(err)
	}

	if subdomain != "custom.example.com" {
		t.Errorf("expected requested subdomain, got %q", subdomain)
	}
}

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTunnel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/organizations/org%2Fone/tunnels" && r.URL.Path != "/api/v1/organizations/org/one/tunnels" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Fatalf("missing authorization")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"tunnel-1"}`))
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL, APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	tunnel, err := client.CreateTunnel(context.Background(), "org/one", map[string]any{"name": "preview"})
	if err != nil {
		t.Fatal(err)
	}
	if tunnel["id"] != "tunnel-1" {
		t.Fatalf("tunnel = %#v", tunnel)
	}
}

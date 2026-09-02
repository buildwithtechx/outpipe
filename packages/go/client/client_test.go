package client

import (
	"context"
	"encoding/json"
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

func TestTunnelLifecycleAgainstHTTPServer(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath())
		if r.Header.Get("Authorization") != "Bearer integration-key" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Method == http.MethodPost {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode request body: %v", err)
			}
			if payload["name"] != "preview" {
				t.Errorf("request payload = %#v", payload)
			}
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/tunnels/tunnel-1" {
			_, _ = w.Write([]byte(`{"id":"tunnel-1","status":"active","publicUrl":"https://preview.outpipe.app"}`))
			return
		}
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":"tunnel-1","status":"active"}]`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"tunnel-1","status":"active","publicUrl":"https://preview.outpipe.app"}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, APIKey: "integration-key"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	created, err := client.CreateTunnel(ctx, "org/one", map[string]any{"name": "preview"})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := client.Tunnels(ctx, "org/one")
	if err != nil {
		t.Fatal(err)
	}
	inspected, err := client.Tunnel(ctx, "tunnel-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.RevokeTunnel(ctx, "tunnel-1"); err != nil {
		t.Fatal(err)
	}

	if created["publicUrl"] != "https://preview.outpipe.app" || len(listed) != 1 || inspected["status"] != "active" {
		t.Fatalf("unexpected lifecycle result: created=%#v listed=%#v inspected=%#v", created, listed, inspected)
	}
	want := []string{
		"POST /api/v1/organizations/org%2Fone/tunnels",
		"GET /api/v1/organizations/org%2Fone/tunnels",
		"GET /api/v1/tunnels/tunnel-1",
		"DELETE /api/v1/tunnels/tunnel-1",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
	for index := range want {
		if requests[index] != want[index] {
			t.Errorf("request[%d] = %q, want %q", index, requests[index], want[index])
		}
	}
}

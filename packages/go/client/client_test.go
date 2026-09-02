package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

type httpTunnelContract struct {
	Authentication string         `json:"authentication"`
	Tunnel         map[string]any `json:"tunnel"`
	Routes         map[string]struct {
		Method  string         `json:"method"`
		Path    string         `json:"path"`
		Request map[string]any `json:"request"`
		Status  int            `json:"status"`
	} `json:"routes"`
}

func loadHTTPContract(t *testing.T) httpTunnelContract {
	t.Helper()
	data, err := os.ReadFile("../../../protocol/fixtures/http_tunnel_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract httpTunnelContract
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	return contract
}

func TestCreateTunnel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/v1/organizations/org%2Fone/tunnels" {
			t.Fatalf("path = %s", r.URL.EscapedPath())
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
	contract := loadHTTPContract(t)
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath())
		if r.Header.Get("Authorization") != contract.Authentication+" integration-key" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path == "/api/v1/organizations/error/401/tunnels" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"authentication required"}`))
			return
		}
		route, ok := contractRoute(contract, r.Method, r.URL.EscapedPath())
		if !ok {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode request body: %v", err)
			}
			if !reflect.DeepEqual(payload, route.Request) {
				t.Errorf("request payload = %#v", payload)
			}
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(route.Status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(route.Status)
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/tunnels/tunnel-1" {
			body, _ := json.Marshal(contract.Tunnel)
			_, _ = w.Write(body)
			return
		}
		if r.Method == http.MethodGet {
			body, _ := json.Marshal([]map[string]any{contract.Tunnel})
			_, _ = w.Write(body)
			return
		}
		body, _ := json.Marshal(contract.Tunnel)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, APIKey: "integration-key"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	created, err := client.CreateTunnel(ctx, "org/one", contract.Routes["create"].Request)
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

	if created["publicHostname"] != "preview.outpipe.app" || len(listed) != 1 || inspected["status"] != "active" {
		t.Fatalf("unexpected lifecycle result: created=%#v listed=%#v inspected=%#v", created, listed, inspected)
	}
	want := []string{contract.Routes["create"].Method + " " + contract.Routes["create"].Path, contract.Routes["list"].Method + " " + contract.Routes["list"].Path, contract.Routes["inspect"].Method + " " + contract.Routes["inspect"].Path, contract.Routes["revoke"].Method + " " + contract.Routes["revoke"].Path}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
	for index := range want {
		if requests[index] != want[index] {
			t.Errorf("request[%d] = %q, want %q", index, requests[index], want[index])
		}
	}
	_, err = client.Tunnels(ctx, "error/401")
	if apiErr, ok := err.(*APIError); !ok || apiErr.StatusCode != http.StatusUnauthorized || apiErr.Message != "authentication required" {
		t.Fatalf("unauthorized error = %#v", err)
	}
}

func contractRoute(contract httpTunnelContract, method, path string) (struct {
	Method  string         `json:"method"`
	Path    string         `json:"path"`
	Request map[string]any `json:"request"`
	Status  int            `json:"status"`
}, bool) {
	for _, route := range contract.Routes {
		if route.Method == method && route.Path == path {
			return route, true
		}
	}
	return struct {
		Method  string         `json:"method"`
		Path    string         `json:"path"`
		Request map[string]any `json:"request"`
		Status  int            `json:"status"`
	}{}, false
}

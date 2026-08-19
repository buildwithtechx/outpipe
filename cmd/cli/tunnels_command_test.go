package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRunTunnelsCommands(t *testing.T) {
	var mu sync.Mutex
	var seen []struct {
		method string
		path   string
		auth   string
		body   string
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := ""

		if r.Body != nil {
			buffer := make([]byte, 4096)
			n, _ := r.Body.Read(buffer)
			body = strings.TrimSpace(string(buffer[:n]))
		}

		mu.Lock()
		seen = append(seen, struct {
			method string
			path   string
			auth   string
			body   string
		}{r.Method, r.URL.Path, r.Header.Get("Authorization"), body})
		mu.Unlock()

		if r.Header.Get("Authorization") != "Bearer key-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/organizations/org-1/tunnels":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"t-new","name":"my-app","protocol":"http","status":"active","publicUrl":"http://my-app.outpipe.app"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/organizations/org-1/tunnels":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"t-1","name":"one","protocol":"http","status":"active","publicUrl":"http://one.outpipe.app"}]`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/tunnels/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"t-1","name":"one","protocol":"http","status":"active","publicUrl":"http://one.outpipe.app"}`))
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/v1/tunnels/"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/tunnels/"):
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_API_KEY", "key-1")

	commands := [][]string{
		{"create", "--organization", "org-1", "my-app"},
		{"list", "--organization", "org-1"},
		{"inspect", "t-1"},
		{"start", "t-1"},
		{"stop", "t-1"},
		{"revoke", "t-1"},
	}

	for _, args := range commands {

		if err := run(args); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(seen) != len(commands) {
		t.Fatalf("expected %d requests, got %d", len(commands), len(seen))
	}

	expected := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/organizations/org-1/tunnels"},
		{http.MethodGet, "/api/v1/organizations/org-1/tunnels"},
		{http.MethodGet, "/api/v1/tunnels/t-1"},
		{http.MethodPatch, "/api/v1/tunnels/t-1/status"},
		{http.MethodPatch, "/api/v1/tunnels/t-1/status"},
		{http.MethodDelete, "/api/v1/tunnels/t-1"},
	}

	for i, want := range expected {

		if seen[i].method != want.method || seen[i].path != want.path {
			t.Errorf("request %d: got %s %s, want %s %s", i, seen[i].method, seen[i].path, want.method, want.path)
		}

		if seen[i].auth != "Bearer key-1" {
			t.Errorf("request %d: expected Bearer key-1, got %q", i, seen[i].auth)
		}
	}

	if !strings.Contains(seen[0].body, `"name":"my-app"`) || !strings.Contains(seen[0].body, `"protocol":"http"`) {
		t.Errorf("create payload missing fields: %s", seen[0].body)
	}

	if !strings.Contains(seen[3].body, `"status":"active"`) {
		t.Errorf("start payload: %s", seen[3].body)
	}

	if !strings.Contains(seen[4].body, `"status":"disconnected"`) {
		t.Errorf("stop payload: %s", seen[4].body)
	}
}

func TestRunTunnelsCommandValidationErrors(t *testing.T) {
	cliTestEnv(t, "http://api.test")
	t.Setenv("OUTPIPE_API_KEY", "key-1")

	for _, args := range [][]string{
		{"list"},
		{"create", "name"},
		{"inspect"},
		{"start"},
		{"revoke"},
		{"bogus"},
	} {

		if err := run(args); err == nil {
			t.Errorf("args %v: expected error", args)
		}
	}
}

func TestRunTunnelsCommandJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"t-1","name":"one","protocol":"http","status":"active","publicUrl":"http://one.outpipe.app"}]`))
	}))
	defer server.Close()
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_API_KEY", "key-1")

	output := captureStdout(t, func() {
		if err := run([]string{"list", "--organization", "org-1", "--json"}); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, `"publicUrl": "http://one.outpipe.app"`) {
		t.Errorf("json output missing tunnel: %s", output)
	}
}

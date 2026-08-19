package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/config"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()

	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer

	var mu sync.Mutex
	output := make([]byte, 0, 4096)
	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 4096)

		for {
			n, readErr := reader.Read(buffer)

			if n > 0 {
				mu.Lock()
				output = append(output, buffer[:n]...)
				mu.Unlock()
			}

			if readErr != nil {
				return
			}
		}
	}()

	fn()
	_ = writer.Close()
	os.Stdout = original
	<-done
	mu.Lock()
	value := string(output)
	mu.Unlock()

	return value
}

func cliTestEnv(t *testing.T, apiURL string) {
	t.Helper()
	t.Setenv("OUTPIPE_API_URL", apiURL)
	t.Setenv("OUTPIPE_RELAY_URL", "ws://relay.test")
	t.Setenv("OUTPIPE_CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))
}

func TestRunDispatchesCommands(t *testing.T) {
	cliTestEnv(t, "http://api.test")

	if err := run([]string{"version"}); err != nil {
		t.Fatalf("version: %v", err)
	}

	if err := run(nil); err != nil {
		t.Fatalf("no args: %v", err)
	}

	if err := run([]string{"help"}); err != nil {
		t.Fatalf("help: %v", err)
	}

	if err := run([]string{"does-not-exist"}); err == nil || !errors.Is(err, errUnknownCommand) {
		t.Fatalf("unknown command: expected errUnknownCommand, got %v", err)
	}
}

func TestRunHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" {
			http.NotFound(w, r)
			return
		}

		if r.Header.Get("Authorization") != "Bearer key-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_API_KEY", "key-1")

	if err := run([]string{"health"}); err != nil {
		t.Fatalf("health: %v", err)
	}

	unreachable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unreachable.Close()
	cliTestEnv(t, unreachable.URL)

	if err := run([]string{"health"}); err == nil {
		t.Fatal("expected health to fail against unavailable server")
	}
}

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

	cfg := config.CLIConfig{APIURL: server.URL, APIKey: "key-1"}
	commands := []struct {
		args []string
	}{
		{[]string{"create", "--organization", "org-1", "my-app"}},
		{[]string{"list", "--organization", "org-1"}},
		{[]string{"inspect", "t-1"}},
		{[]string{"start", "t-1"}},
		{[]string{"stop", "t-1"}},
		{[]string{"revoke", "t-1"}},
	}

	for _, command := range commands {

		if err := runTunnelsCommand(cfg, command.args); err != nil {
			t.Fatalf("%v: %v", command.args, err)
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
	cfg := config.CLIConfig{APIURL: "http://api.test", APIKey: "key-1"}

	for _, args := range [][]string{
		{"list"},
		{"create", "name"},
		{"inspect"},
		{"start"},
		{"revoke"},
		{"bogus"},
	} {

		if err := runTunnelsCommand(cfg, args); err == nil {
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

	cfg := config.CLIConfig{APIURL: server.URL, APIKey: "key-1"}
	output := captureStdout(t, func() {
		if err := runTunnelsCommand(cfg, []string{"list", "--organization", "org-1", "--json"}); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, `"publicUrl": "http://one.outpipe.app"`) {
		t.Errorf("json output missing tunnel: %s", output)
	}
}

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

func TestRunLoginSavesAgentToken(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cliTestEnv(t, "http://api.test")
	t.Setenv("OUTPIPE_CONFIG_PATH", configPath)

	cfg, err := config.LoadCLI()

	if err != nil {
		t.Fatal(err)
	}

	if err := runLogin(cfg, []string{"--agent-token", "token-1", "--api-key", "key-1"}); err != nil {
		t.Fatal(err)
	}

	stored, err := config.LoadCLIFile(configPath)

	if err != nil {
		t.Fatal(err)
	}

	if stored.AgentToken != "token-1" || stored.APIKey != "key-1" {
		t.Errorf("stored credentials wrong: %+v", stored)
	}

	if stored.Version != config.CurrentCLIConfigVersion {
		t.Errorf("expected version %d, got %d", config.CurrentCLIConfigVersion, stored.Version)
	}
}

func TestRunLoginDeviceFlow(t *testing.T) {
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/device/start":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":"ABC-123"}`))
		case "/api/v1/auth/device/poll":
			polls++

			if polls < 2 {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"pending"}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"approved","token":"device-token"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_CONFIG_PATH", configPath)

	cfg, err := config.LoadCLI()

	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- runLogin(cfg, []string{"--code", "ABC-123"}) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("device login did not complete")
	}

	stored, err := config.LoadCLIFile(configPath)

	if err != nil {
		t.Fatal(err)
	}

	if stored.AgentToken != "device-token" {
		t.Errorf("expected device token stored, got %q", stored.AgentToken)
	}

	if polls < 2 {
		t.Errorf("expected at least 2 polls, got %d", polls)
	}
}

func TestRunCompletion(t *testing.T) {

	if err := runCompletion([]string{"--shell", "zsh"}); err != nil {
		t.Fatal(err)
	}

	if err := runCompletion([]string{"--shell", "powershell"}); err == nil {
		t.Fatal("expected unsupported shell error")
	}
}

func TestRunLoadsStoredCredentials(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")

	if err := config.SaveCLI(config.CLIConfig{ConfigPath: configPath, APIKey: "stored-key", AgentToken: "stored-token"}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" && r.Header.Get("Authorization") == "Bearer stored-key" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_CONFIG_PATH", configPath)

	if err := run([]string{"health"}); err != nil {
		t.Fatalf("stored credentials not loaded: %v", err)
	}
}

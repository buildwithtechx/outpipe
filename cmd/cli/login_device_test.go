package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/config"
)

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

	done := make(chan error, 1)
	go func() { done <- run([]string{"login", "--code", "ABC-123"}) }()

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

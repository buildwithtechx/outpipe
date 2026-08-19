package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

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

func TestRunCompletion(t *testing.T) {
	cliTestEnv(t, "http://api.test")

	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {

		if err := run([]string{"completion", shell}); err != nil {
			t.Errorf("completion %s: %v", shell, err)
		}
	}

	if err := run([]string{"completion", "bogus"}); err != nil {
		t.Errorf("completion with unknown shell must fall back to help: %v", err)
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

func TestMainEnvIsolation(t *testing.T) {
	if os.Getenv("OUTPIPE_API_URL") != "" {
		t.Skip("environment already set; skipping isolation check")
	}

	cliTestEnv(t, "http://isolated.test")

	if os.Getenv("OUTPIPE_API_URL") != "http://isolated.test" {
		t.Fatal("cliTestEnv did not set the API URL")
	}
}

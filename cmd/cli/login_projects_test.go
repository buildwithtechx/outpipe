package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoginCredentialsAndEndpointPersistAcrossProjects(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, key := range []string{"OUTPIPE_CONFIG_PATH", "OUTPIPE_API_KEY", "OUTPIPE_AGENT_TOKEN"} {
		t.Setenv(key, "")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" || r.Header.Get("Authorization") != "Bearer project-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("OUTPIPE_API_URL", server.URL)
	t.Setenv("OUTPIPE_RELAY_URL", "ws://relay.test")
	firstProject, secondProject := t.TempDir(), t.TempDir()
	t.Chdir(firstProject)
	if err := run([]string{"login", "--agent-token", "project-agent", "--api-key", "project-key"}); err != nil {
		t.Fatalf("save first-project login: %v", err)
	}
	for _, key := range []string{"OUTPIPE_API_URL", "OUTPIPE_RELAY_URL", "OUTPIPE_API_KEY", "OUTPIPE_AGENT_TOKEN"} {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("clear test override: %v", err)
		}
	}
	if err := os.Chdir(secondProject); err != nil {
		t.Fatalf("switch test project: %v", err)
	}
	if err := run([]string{"health"}); err != nil {
		t.Fatalf("second project did not reuse stored endpoint and credential: %v", err)
	}
	t.Setenv("OUTPIPE_API_URL", "")
	t.Setenv("OUTPIPE_RELAY_URL", "")
	t.Setenv("OUTPIPE_API_KEY", "replace_with_api_key")
	t.Setenv("OUTPIPE_AGENT_TOKEN", "replace_with_agent_token")
	cfg, err := loadCLISettings()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIURL != server.URL || cfg.RelayURL != "ws://relay.test" || cfg.APIKey != "project-key" || cfg.AgentToken != "project-agent" {
		t.Fatal("empty overrides or example placeholders replaced saved login settings")
	}
}

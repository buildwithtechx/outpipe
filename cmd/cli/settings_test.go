package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

func TestLoginPreservesUnsupportedConfigVersion(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("OUTPIPE_CONFIG_PATH", path)
	data, err := json.Marshal(map[string]any{"version": config.CurrentCLIConfigVersion + 1, "futureSetting": "preserve"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"login", "--agent-token", "new-agent", "--api-key", "new-key"}); err == nil {
		t.Fatal("login accepted unsupported saved config version")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, after) {
		t.Fatal("login overwrote unsupported saved settings")
	}
}

func TestLoginRepairsCorruptSavedSettings(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("OUTPIPE_CONFIG_PATH", path)
	t.Setenv("OUTPIPE_API_URL", "https://api.outpipe.dev")
	t.Setenv("OUTPIPE_RELAY_URL", "wss://relay.outpipe.app/v1/connect")
	t.Setenv("OUTPIPE_API_KEY", "")
	t.Setenv("OUTPIPE_AGENT_TOKEN", "")
	if err := os.WriteFile(path, []byte("invalid json"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"login", "--agent-token", "repaired-agent", "--api-key", "repaired-key"}); err != nil {
		t.Fatalf("login could not repair corrupt saved settings: %v", err)
	}
	cfg, err := config.LoadCLIFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentToken != "repaired-agent" || cfg.APIKey != "repaired-key" {
		t.Fatal("login did not replace corrupt settings with new credentials")
	}
}

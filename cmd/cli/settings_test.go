package main

import (
	"os"
	"path/filepath"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

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

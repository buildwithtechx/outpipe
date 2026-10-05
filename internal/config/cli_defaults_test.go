package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLIHostedDefaultsAndUserConfiguration(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, key := range []string{"OUTPIPE_API_URL", "OUTPIPE_RELAY_URL", "OUTPIPE_DOMAIN", "OUTPIPE_CONFIG_PATH", "OUTPIPE_API_KEY", "OUTPIPE_AGENT_TOKEN"} {
		t.Setenv(key, "")
	}
	cfg, err := LoadCLI()
	if err != nil {
		t.Fatalf("load hosted CLI settings: %v", err)
	}
	if cfg.APIURL != "https://api.outpipe.dev" || cfg.RelayURL != "wss://relay.outpipe.app/v1/connect" || cfg.PublicDomain != "outpipe.app" {
		t.Fatal("CLI did not default to hosted endpoints")
	}
	directory, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("find user configuration directory: %v", err)
	}
	if cfg.ConfigPath != filepath.Join(directory, "outpipe", "config.json") {
		t.Fatal("CLI credentials are not stored in the user configuration directory")
	}
	t.Setenv("OUTPIPE_API_URL", "http://localhost:8080")
	t.Setenv("OUTPIPE_RELAY_URL", "ws://localhost:8081")
	t.Setenv("OUTPIPE_DOMAIN", "outpipe.localhost")
	t.Setenv("OUTPIPE_CONFIG_PATH", filepath.Join(t.TempDir(), "local.json"))
	local, err := LoadCLI()
	if err != nil {
		t.Fatalf("load local CLI overrides: %v", err)
	}
	if local.APIURL != "http://localhost:8080" || local.RelayURL != "ws://localhost:8081" || local.PublicDomain != "outpipe.localhost" || local.ConfigPath == cfg.ConfigPath {
		t.Fatal("explicit development overrides were ignored")
	}
}

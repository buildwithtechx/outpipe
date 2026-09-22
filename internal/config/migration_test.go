package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIConfigFileRoundTripWithVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := CLIConfig{ConfigPath: path, APIKey: "key-1", AgentToken: "token-1"}

	if err := SaveCLI(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	stored, err := LoadCLIFile(path)

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if stored.Version != CurrentCLIConfigVersion {
		t.Errorf("expected version %d, got %d", CurrentCLIConfigVersion, stored.Version)
	}

	if stored.APIKey != "key-1" || stored.AgentToken != "token-1" {
		t.Errorf("credentials did not round trip: %+v", stored)
	}
}

func TestCLIConfigLegacyFileMigratesToCurrentVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := map[string]any{"APIKey": "legacy-key", "AgentToken": "legacy-token", "APIURL": "http://legacy.test"}

	data, err := json.Marshal(legacy)

	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadCLIFile(path)

	if err != nil {
		t.Fatalf("load legacy config: %v", err)
	}

	if cfg.Version != CurrentCLIConfigVersion {
		t.Errorf("legacy file must migrate to version %d, got %d", CurrentCLIConfigVersion, cfg.Version)
	}

	if cfg.APIKey != "legacy-key" || cfg.AgentToken != "legacy-token" {
		t.Errorf("legacy credentials not preserved: %+v", cfg)
	}
}

func TestCLIConfigNewerVersionRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data, err := json.Marshal(map[string]any{"Version": CurrentCLIConfigVersion + 1, "AgentToken": "x"})

	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadCLIFile(path); err == nil {
		t.Fatal("expected newer config version to be rejected")
	}
}

func TestCLIConfigCorruptFileRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := os.WriteFile(path, []byte("{ not json"), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadCLIFile(path); err == nil || !strings.Contains(err.Error(), "decode cli config") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestCLIConfigMissingFileReturnsReadError(t *testing.T) {
	if _, err := LoadCLIFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected read error for missing config")
	}
}

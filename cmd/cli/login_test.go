package main

import (
	"path/filepath"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

func TestRunLoginSavesAgentToken(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cliTestEnv(t, "http://api.test")
	t.Setenv("OUTPIPE_CONFIG_PATH", configPath)

	if err := run([]string{"login", "--agent-token", "token-1", "--api-key", "key-1"}); err != nil {
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

func TestRunLoginSavesApiKeyOnly(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cliTestEnv(t, "http://api.test")
	t.Setenv("OUTPIPE_CONFIG_PATH", configPath)

	if err := run([]string{"login", "--agent-token", "token-1"}); err != nil {
		t.Fatal(err)
	}

	stored, err := config.LoadCLIFile(configPath)

	if err != nil {
		t.Fatal(err)
	}

	if stored.AgentToken != "token-1" {
		t.Errorf("expected agent token stored, got %q", stored.AgentToken)
	}
}

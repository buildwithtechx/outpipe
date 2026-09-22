package config

import "testing"

func TestLoadUsesOutpipeEnvironmentPrefix(t *testing.T) {
	t.Setenv("OUTPIPE_PORT", "9090")
	t.Setenv("OUTPIPE_APP_NAME", "test-tunnel")
	t.Setenv("OUTPIPE_DATABASE_MAX_CONNS", "12")
	t.Setenv("OUTPIPE_GOOGLE_CLIENT_ID", "google-client")
	t.Setenv("OUTPIPE_GITHUB_CLIENT_ID", "github-client")
	t.Setenv("OUTPIPE_ZEPTO_API_KEY", "zepto-key")

	cfg, err := LoadAPI()

	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.App.Port)
	}

	if cfg.App.Name != "test-tunnel" {
		t.Fatalf("expected test-tunnel, got %s", cfg.App.Name)
	}

	if cfg.Database.MaxConns != 12 {
		t.Fatalf("expected 12 database connections, got %d", cfg.Database.MaxConns)
	}

	if cfg.Auth.GoogleClientID != "google-client" || cfg.Auth.GitHubClientID != "github-client" {
		t.Fatal("oauth configuration did not load")
	}

	if cfg.Mail.ZeptoAPIKey != "zepto-key" {
		t.Fatal("zepto configuration did not load")
	}
}

func TestLoadRelayDoesNotRequireDatabase(t *testing.T) {
	t.Setenv("OUTPIPE_PORT", "8081")
	t.Setenv("OUTPIPE_MAX_CONNECTIONS", "20")

	cfg, err := LoadRelay()

	if err != nil {
		t.Fatalf("load relay config: %v", err)
	}

	if cfg.App.Port != "8081" {
		t.Fatalf("expected relay port 8081, got %s", cfg.App.Port)
	}
}

func TestLoadCLIDoesNotRequireServerSecrets(t *testing.T) {
	cfg, err := LoadCLI()

	if err != nil {
		t.Fatalf("load cli config: %v", err)
	}

	if cfg.APIKey != "" {
		t.Fatal("expected API key to be optional")
	}
}

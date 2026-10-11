package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestTunnelTLSSettingEnablesRelayTLS(t *testing.T) {
	t.Setenv("OUTPIPE_TUNNEL_REQUIRE_TLS", "true")
	t.Setenv("OUTPIPE_TLS_CERT_FILE", "certificate.pem")
	t.Setenv("OUTPIPE_TLS_KEY_FILE", "key.pem")
	cfg, err := LoadRelay()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.App.RequireTLS {
		t.Fatal("tunnel TLS setting was ignored")
	}
}

func TestLoadAPIFromYAMLDevelopment(t *testing.T) {
	t.Setenv("OUTPIPE_ENV", "development")
	t.Setenv("ENV", "development")
	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("load api config: %v", err)
	}
	if cfg.App.Environment != "development" {
		t.Fatalf("expected development environment, got %s", cfg.App.Environment)
	}
	if cfg.Agent.DefaultProvider != "mock" {
		t.Fatalf("expected mock agent provider from development yaml, got %s", cfg.Agent.DefaultProvider)
	}
}

func TestLoadAPIFromYAMLProduction(t *testing.T) {
	t.Setenv("OUTPIPE_ENV", "production")
	t.Setenv("ENV", "production")
	t.Setenv("OUTPIPE_DATABASE_URL", "postgres://prod_user:prod_pass@db.internal:5432/outpipe_prod?sslmode=require")
	t.Setenv("OUTPIPE_AUTH_ENCRYPTION_KEY", "12345678901234567890123456789012")
	t.Setenv("OUTPIPE_INTERNAL_API_SECRET", "super-secret-service-token-12345")
	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("load api production config: %v", err)
	}
	if cfg.App.Environment != "production" {
		t.Fatalf("expected production environment, got %s", cfg.App.Environment)
	}
	if !cfg.App.RequireTLS {
		t.Fatal("expected production yaml to require TLS")
	}
	if cfg.Tunnel.MaxConnections != 50000 {
		t.Fatalf("expected 50000 max connections in production yaml, got %d", cfg.Tunnel.MaxConnections)
	}
}

func TestLoadAPIFromExplicitConfigPath(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "custom.yaml")
	customYAML := `
app:
  port: "9999"
  name: "custom-outpipe"
database:
  max_conns: 88
tunnel:
  domain: "custom.domain.dev"
agent:
  default_provider: "anthropic"
`
	if err := os.WriteFile(tempFile, []byte(customYAML), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OUTPIPE_CONFIG_PATH", tempFile)
	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("load explicit yaml config: %v", err)
	}
	if cfg.App.Port != "9999" {
		t.Fatalf("expected port 9999 from custom yaml, got %s", cfg.App.Port)
	}
	if cfg.App.Name != "custom-outpipe" {
		t.Fatalf("expected name custom-outpipe, got %s", cfg.App.Name)
	}
	if cfg.Database.MaxConns != 88 {
		t.Fatalf("expected max_conns 88, got %d", cfg.Database.MaxConns)
	}
	if cfg.Agent.DefaultProvider != "anthropic" {
		t.Fatalf("expected anthropic agent provider, got %s", cfg.Agent.DefaultProvider)
	}
}

func TestEnvironmentOverridesYAMLValues(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "base.yaml")
	baseYAML := `
app:
  port: "7000"
  name: "yaml-base"
database:
  max_conns: 20
`
	if err := os.WriteFile(tempFile, []byte(baseYAML), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OUTPIPE_CONFIG_PATH", tempFile)
	t.Setenv("OUTPIPE_PORT", "7777")
	t.Setenv("OUTPIPE_DATABASE_MAX_CONNS", "99")

	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.Port != "7777" {
		t.Fatalf("expected env override port 7777, got %s", cfg.App.Port)
	}
	if cfg.Database.MaxConns != 99 {
		t.Fatalf("expected env override max_conns 99, got %d", cfg.Database.MaxConns)
	}
	if cfg.App.Name != "yaml-base" {
		t.Fatalf("expected name yaml-base from yaml file, got %s", cfg.App.Name)
	}
}

func TestLoadRelayFromYAMLDevelopment(t *testing.T) {
	t.Setenv("OUTPIPE_ENV", "development")
	t.Setenv("ENV", "development")
	cfg, err := LoadRelay()
	if err != nil {
		t.Fatalf("load relay config: %v", err)
	}
	if cfg.RelayID != "relay-local" {
		t.Fatalf("expected relay-local from yaml, got %s", cfg.RelayID)
	}
	if cfg.Tunnel.Domain != "outpipe.localhost" {
		t.Fatalf("expected outpipe.localhost from yaml, got %s", cfg.Tunnel.Domain)
	}
}

func TestLoadCronFromYAMLDevelopment(t *testing.T) {
	t.Setenv("OUTPIPE_ENV", "development")
	t.Setenv("ENV", "development")
	t.Setenv("OUTPIPE_DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	cfg, err := LoadCron()
	if err != nil {
		t.Fatalf("load cron config: %v", err)
	}
	if cfg.Backup.Keep != 7 {
		t.Fatalf("expected backup keep 7 from yaml, got %d", cfg.Backup.Keep)
	}
	if cfg.Backup.Directory != "data/backups" {
		t.Fatalf("expected data/backups directory from yaml, got %s", cfg.Backup.Directory)
	}
	if cfg.Metrics {
		t.Fatal("expected metrics false in development yaml")
	}
}

func TestLoadCronFromYAMLProduction(t *testing.T) {
	t.Setenv("OUTPIPE_ENV", "production")
	t.Setenv("ENV", "production")
	t.Setenv("OUTPIPE_DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	cfg, err := LoadCron()
	if err != nil {
		t.Fatalf("load cron production config: %v", err)
	}
	if cfg.Backup.Keep != 30 {
		t.Fatalf("expected backup keep 30 from production yaml, got %d", cfg.Backup.Keep)
	}
	if cfg.Backup.Directory != "/var/backups/outpipe" {
		t.Fatalf("expected /var/backups/outpipe from production yaml, got %s", cfg.Backup.Directory)
	}
	if !cfg.Metrics {
		t.Fatal("expected metrics true in production yaml")
	}
}

func TestLoadCheckFromYAMLDevelopment(t *testing.T) {
	t.Setenv("OUTPIPE_ENV", "development")
	t.Setenv("ENV", "development")
	cfg, err := LoadCheck()
	if err != nil {
		t.Fatalf("load check config: %v", err)
	}
	if cfg.RelayURL != "http://127.0.0.1:8081" {
		t.Fatalf("expected relay url http://127.0.0.1:8081 from yaml, got %s", cfg.RelayURL)
	}
	if cfg.App.Port != "8082" {
		t.Fatalf("expected port 8082 from yaml, got %s", cfg.App.Port)
	}
}

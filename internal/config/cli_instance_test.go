package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIConfigFilenameKeepsLegacyForHosted(t *testing.T) {
	if got := CLIConfigFilename(""); got != "config.json" {
		t.Fatalf("empty origin must stay legacy, got %s", got)
	}
	for _, origin := range []string{"https://api.outpipe.dev", "https://API.OUTPIPE.DEV:443/", "https://api.outpipe.dev."} {
		if got := CLIConfigFilename(origin); got != "config.json" {
			t.Fatalf("hosted origin %s must stay legacy, got %s", origin, got)
		}
	}
}

func TestCLIConfigFilenameScopesCustomOrigins(t *testing.T) {
	filename := CLIConfigFilename("https://console.example.test")
	if !strings.HasPrefix(filename, "config.instance.") || !strings.HasSuffix(filename, ".json") || len(filename) != len("config.instance.")+64+len(".json") {
		t.Fatalf("custom origin must map to an instance hash file, got %s", filename)
	}
	if other := CLIConfigFilename("https://CONSOLE.EXAMPLE.TEST:443/"); other != filename {
		t.Fatal("canonical equivalent origins must share one instance file")
	}
	if other := CLIConfigFilename("https://other.example.test"); other == filename {
		t.Fatal("distinct origins must not share an instance file")
	}
	if other := CLIConfigFilename("https://console.example.test:8443"); other == filename {
		t.Fatal("distinct ports must not share an instance file")
	}
	if filename != filepath.Base(filename) {
		t.Fatal("instance filename must not contain a path")
	}
}

func TestCanonicalAPIOriginRequiresTLSExceptLoopback(t *testing.T) {
	for _, origin := range []string{"http://localhost:8080", "http://127.0.0.1:8080", "http://127.5.6.7:8080", "http://[::1]:8080"} {
		if _, err := CanonicalAPIOrigin(origin); err != nil {
			t.Fatalf("loopback origin %s rejected: %v", origin, err)
		}
	}
	for _, origin := range []string{"http://192.168.1.5", "http://console.example.test", "ftp://console.example.test", "https://name:secret@console.example.test", "https://console.example.test/path", "https://console.example.test/?query=1", "https://console.example.test/#key", "not-a-url"} {
		if _, err := CanonicalAPIOrigin(origin); err == nil {
			t.Fatalf("origin %s must be rejected", origin)
		}
	}
}

func TestLoadCLIDefaultsToInstanceFile(t *testing.T) {
	t.Setenv("OUTPIPE_API_URL", "https://console.example.test")
	t.Setenv("OUTPIPE_RELAY_URL", "wss://relay.example.test/v1/connect")
	t.Setenv("OUTPIPE_CONFIG_PATH", "")
	cfg, err := LoadCLI()
	if err != nil {
		t.Fatalf("load CLI with custom origin: %v", err)
	}
	if base := filepath.Base(cfg.ConfigPath); base != CLIConfigFilename("https://console.example.test") {
		t.Fatalf("expected instance config file, got %s", cfg.ConfigPath)
	}
}

func TestLoadCLIRejectsInsecureRemoteOrigin(t *testing.T) {
	t.Setenv("OUTPIPE_API_URL", "http://192.168.1.5:8080")
	t.Setenv("OUTPIPE_CONFIG_PATH", "")
	if _, err := LoadCLI(); err == nil {
		t.Fatal("insecure remote API origin must be rejected")
	}
}

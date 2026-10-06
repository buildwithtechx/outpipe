package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

func TestMachineTunnelValidatesOriginBeforeCreatingTunnel(t *testing.T) {
	invalidCA := filepath.Join(t.TempDir(), "invalid.pem")
	if err := os.WriteFile(invalidCA, []byte("invalid certificate"), 0600); err != nil {
		t.Fatalf("write invalid CA: %v", err)
	}
	for _, test := range []struct {
		name    string
		flags   []string
		message string
	}{
		{"TLS flags with HTTP", []string{"--origin-server-name", "localhost"}, "require --protocol https"},
		{"missing CA", []string{"--protocol", "https", "--origin-ca", filepath.Join(t.TempDir(), "missing.pem")}, "read origin CA file"},
		{"malformed CA", []string{"--protocol", "https", "--origin-ca", invalidCA}, "no valid PEM certificates"},
		{"invalid server name", []string{"--protocol", "https", "--origin-server-name", "https://localhost"}, "server name must be"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("OUTPIPE_TOKEN", "test-machine-token")
			var called atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called.Store(true)
				w.WriteHeader(http.StatusBadRequest)
			}))
			defer server.Close()
			command := newMachineTunnelCommand(config.CLIConfig{APIURL: server.URL})
			command.SetArgs(append([]string{"preview"}, test.flags...))
			err := command.Execute()
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("expected origin validation error, got %v", err)
			}
			if called.Load() {
				t.Fatal("invalid origin configuration created a remote tunnel")
			}
		})
	}
}

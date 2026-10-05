package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"outpipe.dev/outpipe/internal/config"
)

func TestMachineTunnelUsesSharedScopedToken(t *testing.T) {
	for _, test := range []struct{ name, shared, dedicated, want string }{
		{"shared token", "shared-token", "", "shared-token"},
		{"dedicated token", "", "tunnel-token", "tunnel-token"},
		{"dedicated override", "secrets-token", "tunnel-token", "tunnel-token"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("OUTPIPE_TOKEN", test.shared)
			t.Setenv("OUTPIPE_MACHINE_TOKEN", test.dedicated)
			want := "Bearer " + test.want
			var called atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called.Store(true)
				if r.Header.Get("Authorization") != want || r.URL.Path != "/api/v1/machines/tunnels" {
					t.Error("machine request used incorrect credential or endpoint")
				}
				w.WriteHeader(http.StatusBadRequest)
			}))
			defer server.Close()
			command := newMachineTunnelCommand(config.CLIConfig{APIURL: server.URL})
			command.SetArgs([]string{"preview"})
			if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "status 400") || !called.Load() {
				t.Fatal("machine command did not call the API with a scoped token")
			}
		})
	}
}

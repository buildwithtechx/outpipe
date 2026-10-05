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
	for _, shared := range []bool{true, false} {
		t.Run(map[bool]string{true: "shared token", false: "legacy alias"}[shared], func(t *testing.T) {
			t.Setenv("OUTPIPE_TOKEN", "")
			t.Setenv("OUTPIPE_MACHINE_TOKEN", "legacy-token")
			want := "Bearer legacy-token"
			if shared {
				t.Setenv("OUTPIPE_TOKEN", "shared-token")
				want = "Bearer shared-token"
			}
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

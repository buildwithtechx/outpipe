package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

func TestManagedPolicyCacheLimitsCallsAndRefreshesRevocation(t *testing.T) {
	var calls atomic.Int32
	var revoked atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		status := "active"
		if revoked.Load() {
			status = "revoked"
		}
		if _, err := w.Write([]byte(`{"organizationId":"org","machineTokenId":"machine","status":"` + status + `"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	resolver, err := NewInternalTunnelResolver(server.URL, "secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	for range 100 {
		if _, err := resolver.Resolve(context.Background(), "tunnel"); err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("policy calls = %d", calls.Load())
	}
	revoked.Store(true)
	resolver.mu.Lock()
	entry := resolver.policies["tunnel"]
	entry.expires = time.Now().Add(-time.Second)
	resolver.policies["tunnel"] = entry
	resolver.mu.Unlock()
	policy, err := resolver.Resolve(context.Background(), "tunnel")
	if err != nil || policy.Status != "revoked" || calls.Load() != 2 {
		t.Fatalf("revocation was not refreshed: %v", err)
	}
}

func TestManagedPolicyRejectsConflictingHostnames(t *testing.T) {
	handler := &Handler{}
	_, err := handler.resolveManagedPolicy(context.Background(), AgentIdentity{}, &protocol.OpenTunnel{Subdomain: "owned", CustomDomain: "foreign.example.com"})
	if err == nil {
		t.Fatal("conflicting hostname precedence accepted")
	}
}

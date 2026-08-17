package relay

import (
	"context"
	"fmt"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/pkg/protocol"
)

type mockAuthenticator struct {
	tokens map[string]AgentIdentity
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, token string) (AgentIdentity, error) {

	if id, ok := m.tokens[token]; ok {
		return id, nil
	}

	return AgentIdentity{}, fmt.Errorf("invalid token")
}

func dummySend(ctx context.Context, envelope protocol.Envelope) error {
	return nil
}

func TestCrossTunnelAccessIsolation(t *testing.T) {
	auth := &mockAuthenticator{tokens: map[string]AgentIdentity{
		"token-org-1": {AgentID: "agent-1", OrganizationID: "org-1"},
		"token-org-2": {AgentID: "agent-2", OrganizationID: "org-2"},
	}}
	sessions := engine.NewSessionRegistry()
	router, err := engine.NewRequestRouter(sessions, 10*time.Second)

	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	tcp := NewTCPManager()
	udp := NewUDPManager()

	handler, err := NewHandler(auth, sessions, router, tcp, udp, 10)

	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	session1 := engine.Session{ID: "sess-1", OrganizationID: "org-1", TunnelID: "tunnel-org-1", Send: dummySend}

	if err := sessions.Reserve(session1, false); err != nil {
		t.Fatalf("reserve session 1: %v", err)
	}

	closeMsg, _ := protocol.EncodePayload(protocol.MessageTypeCloseTunnel, "", protocol.CloseTunnel{TunnelID: "tunnel-org-1"})
	envelope, _ := protocol.Decode(closeMsg)

	// Org-2 attempting to close Org-1's tunnel must be rejected
	err = handler.handleMessage(context.Background(), nil, AgentIdentity{AgentID: "agent-2", OrganizationID: "org-2"}, envelope, make(map[string]string))

	if err == nil {
		t.Fatal("expected cross-tunnel modification by org-2 to be rejected, got nil error")
	}

	// Session must still exist in registry

	if _, ok := sessions.Get("tunnel-org-1"); !ok {
		t.Fatal("tunnel-org-1 should not have been deleted by org-2")
	}
}

func TestSessionRegistryReservation(t *testing.T) {
	sessions := engine.NewSessionRegistry()
	session := engine.Session{ID: "sess-1", OrganizationID: "org-1", TunnelID: "tunnel-1", Send: dummySend}

	if err := sessions.Reserve(session, false); err != nil {
		t.Fatalf("reserve session: %v", err)
	}

	if _, ok := sessions.Get("tunnel-1"); !ok {
		t.Fatal("expected tunnel-1 to be present")
	}

	if !sessions.Remove("tunnel-1", "sess-1") {
		t.Fatal("expected tunnel-1 to be removed")
	}
}

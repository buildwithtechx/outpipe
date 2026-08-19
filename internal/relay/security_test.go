package relay

import (
	"context"
	"fmt"
	"net"
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

func newLimitHandler(t *testing.T, orgID, tunnelID string, limit int) (*Handler, *engine.RequestRouter) {
	t.Helper()
	auth := &mockAuthenticator{}
	sessions := engine.NewSessionRegistry()
	router, err := engine.NewRequestRouter(sessions, 10*time.Second)

	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	if err := sessions.Reserve(engine.Session{ID: "sess-1", OrganizationID: orgID, TunnelID: tunnelID, Send: dummySend}, false); err != nil {
		t.Fatalf("reserve session: %v", err)
	}

	handler, err := NewHandler(auth, sessions, router, NewTCPManager(), NewUDPManager(), 10)

	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	handler.setOrganizationLimit(orgID, limit)
	return handler, router
}

func orgConnectionCount(h *Handler, orgID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.orgConnections[orgID]
}

func TestOrganizationConnectionLimitReservation(t *testing.T) {
	handler, _ := newLimitHandler(t, "org-1", "limit-tunnel", 2)

	for i := 0; i < 2; i++ {

		if !handler.allowConnection("limit-tunnel") {
			t.Fatalf("connection %d should be admitted", i+1)
		}
	}

	if handler.allowConnection("limit-tunnel") {
		t.Fatal("connection at the plan limit must be rejected")
	}

	if got := orgConnectionCount(handler, "org-1"); got != 2 {
		t.Fatalf("expected 2 reserved connections, got %d", got)
	}

	handler.updateOrganizationConnections("org-1", -1)
	handler.updateOrganizationConnections("org-1", -1)

	if !handler.allowConnection("limit-tunnel") {
		t.Fatal("capacity must be released after connections close")
	}

	if got := orgConnectionCount(handler, "org-1"); got != 1 {
		t.Fatalf("expected 1 reserved connection after release, got %d", got)
	}
}

func TestTCPAdmissionEnforcesOrganizationLimit(t *testing.T) {
	handler, _ := newLimitHandler(t, "org-1", "admit-tunnel", 2)
	port, err := handler.tcp.Open("admit-tunnel", dummySend)

	if err != nil {
		t.Fatalf("open public listener: %v", err)
	}

	address := net.JoinHostPort("127.0.0.1", fmt.Sprint(port))
	accepted := make([]net.Conn, 0, 2)

	for i := 0; i < 2; i++ {
		connection, dialErr := net.Dial("tcp", address)

		if dialErr != nil {
			t.Fatalf("dial %d: %v", i+1, dialErr)
		}

		accepted = append(accepted, connection)
	}

	third, dialErr := net.Dial("tcp", address)

	if dialErr != nil {
		t.Fatalf("dial 3: %v", dialErr)
	}

	_ = third.Close()
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {

		if got := orgConnectionCount(handler, "org-1"); got == 2 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if got := orgConnectionCount(handler, "org-1"); got != 2 {
		t.Fatalf("expected exactly 2 connections counted, got %d", got)
	}

	handler.tcp.CloseTunnel("admit-tunnel")
	deadline = time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {

		if got := orgConnectionCount(handler, "org-1"); got == 0 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if got := orgConnectionCount(handler, "org-1"); got != 0 {
		t.Fatalf("organization connection accounting leaked after tunnel close: %d", got)
	}

	for _, connection := range accepted {
		_ = connection.Close()
	}
}

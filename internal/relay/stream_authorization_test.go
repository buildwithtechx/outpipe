package relay

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/pkg/protocol"
)

type testMachineResolver struct{ policy ManagedTunnelPolicy }

func TestAuthenticationCannotReplaceExistingCredentialIdentity(t *testing.T) {
	original := AgentIdentity{OrganizationID: "org", AgentID: "agent", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	replacement := original
	replacement.ExpiresAt = time.Now().Add(time.Minute).Unix()
	handler := &Handler{authenticator: &mockAuthenticator{tokens: map[string]AgentIdentity{"replacement": replacement}}}
	state := &connectionState{negotiated: true, authenticated: true, identity: original}
	raw, err := protocol.EncodePayload(protocol.MessageTypeAuth, "auth", protocol.AuthRequest{Token: "replacement"})
	if err != nil {
		t.Fatal(err)
	}
	message, err := protocol.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := handler.handleAuthentication(context.Background(), nil, message, state); err == nil {
		t.Fatal("replacement identity accepted")
	}
	if state.identity != original {
		t.Fatal("connection identity changed")
	}
}

func (r *testMachineResolver) Resolve(context.Context, string) (ManagedTunnelPolicy, error) {
	return r.policy, nil
}

func TestStreamRequiresConnectionOwnerAndActiveMachine(t *testing.T) {
	sessions := engine.NewSessionRegistry()
	if err := sessions.Reserve(engine.Session{ID: "session", TunnelID: "tunnel", OrganizationID: "org", Send: func(context.Context, protocol.Envelope) error { return nil }}, false); err != nil {
		t.Fatal(err)
	}
	machineID := "machine"
	resolver := &testMachineResolver{policy: ManagedTunnelPolicy{OrganizationID: "org", Status: "active", MachineTokenID: &machineID}}
	handler := &Handler{sessions: sessions, managedTunnels: resolver}
	identity := AgentIdentity{OrganizationID: "org", TunnelID: "tunnel", MachineTokenID: machineID, ExpiresAt: time.Now().Add(time.Minute).Unix()}
	owned := map[string]string{"tunnel": "session"}
	if err := handler.authorizeStream(context.Background(), identity, "tunnel", owned); err != nil {
		t.Fatal(err)
	}
	if err := handler.authorizeStream(context.Background(), identity, "tunnel", map[string]string{"tunnel": "other-session"}); err == nil {
		t.Fatal("another connection accessed stream")
	}
	resolver.policy.Status = "revoked"
	if err := handler.authorizeStream(context.Background(), identity, "tunnel", owned); err == nil {
		t.Fatal("revoked machine accessed stream")
	}
	resolver.policy.Status = "active"
	identity.ExpiresAt = time.Now().Add(-time.Minute).Unix()
	if err := handler.authorizeStream(context.Background(), identity, "tunnel", owned); err == nil {
		t.Fatal("expired machine accessed stream")
	}
}

func TestUDPResponseCannotConsumeAnotherTunnelPacket(t *testing.T) {
	manager := NewUDPManager()
	manager.packets["packet"] = udpPacket{tunnelID: "owner"}
	if err := manager.Write("attacker", protocol.UDPResponse{TunnelID: "attacker", PacketID: "packet"}); err == nil {
		t.Fatal("foreign packet accepted")
	}
	if _, exists := manager.packets["packet"]; !exists {
		t.Fatal("foreign packet was consumed")
	}
}

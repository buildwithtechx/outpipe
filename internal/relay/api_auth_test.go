package relay

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/auth"
)

func TestInternalAgentAuthenticator_EdgeJWT(t *testing.T) {
	secret := "test-shared-relay-secret-1234567890"
	authenticator, err := NewInternalAgentAuthenticator("http://localhost:8080", secret, nil)
	if err != nil {
		t.Fatalf("NewInternalAgentAuthenticator failed: %v", err)
	}

	claims := auth.RelayClaims{
		Sub:            "agent_abc123",
		Org:            "org_xyz789",
		MaxTunnels:     10,
		MaxConnections: 50,
		BandwidthBytes: 1024 * 1024 * 500,
		Exp:            time.Now().Add(10 * time.Minute).Unix(),
	}

	token, err := auth.SignRelayToken(claims, secret)
	if err != nil {
		t.Fatalf("SignRelayToken failed: %v", err)
	}

	identity, err := authenticator.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("Authenticate with valid edge JWT failed: %v", err)
	}

	if identity.AgentID != claims.Sub {
		t.Errorf("expected AgentID %s, got %s", claims.Sub, identity.AgentID)
	}
	if identity.OrganizationID != claims.Org {
		t.Errorf("expected OrganizationID %s, got %s", claims.Org, identity.OrganizationID)
	}
	if identity.MaxTunnels != claims.MaxTunnels {
		t.Errorf("expected MaxTunnels %d, got %d", claims.MaxTunnels, identity.MaxTunnels)
	}
}

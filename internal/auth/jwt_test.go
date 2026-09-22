package auth

import (
	"testing"
	"time"
)

func TestSignAndVerifyRelayToken(t *testing.T) {
	secret := "test-secret-key-123456789012345678"
	claims := RelayClaims{
		Sub:            "agent_12345",
		Org:            "org_67890",
		MaxTunnels:     5,
		MaxConnections: 20,
		BandwidthBytes: 1024 * 1024 * 100,
		Exp:            time.Now().Add(5 * time.Minute).Unix(),
	}

	token, err := SignRelayToken(claims, secret)
	if err != nil {
		t.Fatalf("SignRelayToken failed: %v", err)
	}

	verified, err := VerifyRelayToken(token, secret)
	if err != nil {
		t.Fatalf("VerifyRelayToken failed: %v", err)
	}

	if verified.Sub != claims.Sub {
		t.Errorf("expected Sub %s, got %s", claims.Sub, verified.Sub)
	}
	if verified.Org != claims.Org {
		t.Errorf("expected Org %s, got %s", claims.Org, verified.Org)
	}
	if verified.MaxTunnels != claims.MaxTunnels {
		t.Errorf("expected MaxTunnels %d, got %d", claims.MaxTunnels, verified.MaxTunnels)
	}
}

func TestVerifyRelayTokenExpired(t *testing.T) {
	secret := "test-secret-key-123456789012345678"
	claims := RelayClaims{
		Sub: "agent_12345",
		Org: "org_67890",
		Exp: time.Now().Add(-5 * time.Minute).Unix(),
	}

	token, err := SignRelayToken(claims, secret)
	if err != nil {
		t.Fatalf("SignRelayToken failed: %v", err)
	}

	_, err = VerifyRelayToken(token, secret)
	if err == nil {
		t.Fatal("expected expired token to fail verification")
	}
}

func TestVerifyRelayTokenTampered(t *testing.T) {
	secret := "test-secret-key-123456789012345678"
	claims := RelayClaims{
		Sub: "agent_12345",
		Org: "org_67890",
		Exp: time.Now().Add(5 * time.Minute).Unix(),
	}

	token, err := SignRelayToken(claims, secret)
	if err != nil {
		t.Fatalf("SignRelayToken failed: %v", err)
	}

	// Tamper with wrong secret
	_, err = VerifyRelayToken(token, "wrong-secret-key-00000000000000")
	if err == nil {
		t.Fatal("expected tampered token to fail verification")
	}
}

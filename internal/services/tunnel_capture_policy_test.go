package services

import (
	"context"
	"testing"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func setupCapturePolicyService(t *testing.T) (*TunnelService, string) {
	t.Helper()
	db := reviewTestDatabase(t, &models.Tunnel{}, &models.CapturePolicy{})
	tunnelRepo, err := repositories.NewTunnelRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	policyRepo, err := repositories.NewCapturePolicyRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewTunnelService(tunnelRepo)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetCapturePolicies(policyRepo)
	tunnel := &models.Tunnel{Base: models.Base{ID: "tunnel-1"}, OrganizationID: "org", Name: "web", Protocol: models.TunnelProtocolHTTPS, Status: models.TunnelStatusActive, TargetHost: "127.0.0.1", TargetPort: 3000, PublicHostname: "web.example.com", CaptureEnabled: true}
	if err := tunnelRepo.Create(context.Background(), tunnel); err != nil {
		t.Fatal(err)
	}
	return svc, tunnel.ID
}

func TestSetCapturePolicyValidatesRules(t *testing.T) {
	svc, tunnelID := setupCapturePolicyService(t)
	ctx := context.Background()
	if _, err := svc.SetCapturePolicy(ctx, tunnelID, `[{"id":"x","methods":["GET"],"pathPrefix":"/","capture":true}]`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetCapturePolicy(ctx, tunnelID, `not json`); err == nil {
		t.Fatal("malformed rules accepted")
	}
	if _, err := svc.SetCapturePolicy(ctx, tunnelID, `[{"id":"x","methods":["GET"],"pathPrefix":"/","capture":true},{"id":"x","methods":["GET"],"pathPrefix":"/","capture":true}]`); err == nil {
		t.Fatal("duplicate rule ids accepted")
	}
	policy, err := svc.GetCapturePolicy(ctx, tunnelID)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Version != 1 {
		t.Fatalf("policy version = %d", policy.Version)
	}
	allowed, err := svc.CaptureAllowed(ctx, tunnelID, "GET", "/")
	if err != nil || !allowed {
		t.Fatalf("allowed request denied: %v", allowed)
	}
	allowed, err = svc.CaptureAllowed(ctx, tunnelID, "POST", "/")
	if err != nil || allowed {
		t.Fatal("denied request allowed")
	}
}

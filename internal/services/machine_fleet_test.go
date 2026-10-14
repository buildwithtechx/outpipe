package services

import (
	"context"
	"testing"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func setupFleetService(t *testing.T) (*MachineFleetService, context.Context) {
	t.Helper()
	db := reviewTestDatabase(t, &models.SecretMachineToken{}, &models.Tunnel{}, &models.MachineFleet{}, &models.MachineFleetToken{}, &models.MachineFleetTunnel{})
	secretRepo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := NewSecretService(secretRepo, "test-encryption-key-that-is-long-enough")
	if err != nil {
		t.Fatal(err)
	}
	tunnelRepo, err := repositories.NewTunnelRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tunnels, err := NewTunnelService(tunnelRepo)
	if err != nil {
		t.Fatal(err)
	}
	fleetRepo, err := repositories.NewMachineFleetRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	fleetService, err := NewMachineFleetService(fleetRepo, secrets, tunnels)
	if err != nil {
		t.Fatal(err)
	}
	return fleetService, context.Background()
}

func TestFleetBulkCreateRotateAndRevoke(t *testing.T) {
	service, ctx := setupFleetService(t)
	fleet, err := service.CreateFleet(ctx, "org-1", "edge-nodes", "edge fleet")
	if err != nil {
		t.Fatal(err)
	}
	token, err := service.secrets.CreateMachineToken(ctx, "org-1", "", "", "fleet-token", []string{"tunnels:write"}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AttachToken(ctx, "org-1", fleet.ID, token.Token.ID, "primary"); err != nil {
		t.Fatal(err)
	}
	specs := []FleetTunnelSpec{
		{Name: "node-a", Protocol: "http", TargetHost: "127.0.0.1", TargetPort: 3000},
		{Name: "node-b", Protocol: "http", TargetHost: "127.0.0.1", TargetPort: 3001},
	}
	created, err := service.BulkCreateTunnels(ctx, "org-1", fleet.ID, "user-1", specs)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 2 {
		t.Fatalf("created = %d", len(created))
	}
	for _, tunnel := range created {
		if !tunnel.MachineOwned || tunnel.MachineTokenID == nil || *tunnel.MachineTokenID != token.Token.ID {
			t.Fatalf("tunnel not machine-owned: %+v", tunnel)
		}
	}
	status, err := service.Status(ctx, "org-1", fleet.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Tokens) != 1 || len(status.DesiredTunnels) != 2 || len(status.LiveTunnels) != 2 {
		t.Fatalf("status = %+v", status)
	}
	if len(status.Drift) != 0 {
		t.Fatalf("unexpected drift: %+v", status.Drift)
	}

	rotated, err := service.RotateFleetTokens(ctx, "org-1", fleet.ID, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rotated) != 1 || rotated[0].Raw == "" {
		t.Fatalf("rotated = %+v", rotated)
	}
	if rotated[0].Token.ID == token.Token.ID {
		t.Fatal("rotation reused the old token id")
	}
	old, err := service.secrets.ListMachineTokens(ctx, "org-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range old {
		if item.ID == token.Token.ID && item.RevokedAt == nil {
			t.Fatal("old token was not revoked")
		}
	}
	if _, err := service.secrets.VerifyMachineToken(ctx, token.Raw); err == nil {
		t.Fatal("revoked token still verifies")
	}

	if err := service.RevokeFleet(ctx, "org-1", fleet.ID); err != nil {
		t.Fatal(err)
	}
	after, err := service.Status(ctx, "org-1", fleet.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, tunnel := range after.LiveTunnels {
		if tunnel.Status != models.TunnelStatusRevoked {
			t.Fatalf("tunnel not revoked: %+v", tunnel)
		}
	}
}

func TestFleetReconcileReactivatesDrift(t *testing.T) {
	service, ctx := setupFleetService(t)
	fleet, err := service.CreateFleet(ctx, "org-1", "edge-nodes", "")
	if err != nil {
		t.Fatal(err)
	}
	token, err := service.secrets.CreateMachineToken(ctx, "org-1", "", "", "fleet-token", []string{"tunnels:write"}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AttachToken(ctx, "org-1", fleet.ID, token.Token.ID, "primary"); err != nil {
		t.Fatal(err)
	}
	created, err := service.BulkCreateTunnels(ctx, "org-1", fleet.ID, "user-1", []FleetTunnelSpec{{Name: "node-a", Protocol: "http", TargetHost: "127.0.0.1", TargetPort: 3000}})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.tunnels.SetStatus(ctx, created[0].ID, models.TunnelStatusDisconnected); err != nil {
		t.Fatal(err)
	}
	status, err := service.Status(ctx, "org-1", fleet.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Drift) != 1 {
		t.Fatalf("drift = %+v", status.Drift)
	}
	result, err := service.Reconcile(ctx, "org-1", fleet.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Reactivated != 1 {
		t.Fatalf("reconcile result = %+v", result)
	}
	tunnel, err := service.tunnels.Find(ctx, created[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if tunnel.Status != models.TunnelStatusActive {
		t.Fatalf("tunnel status = %s", tunnel.Status)
	}
}

func TestFleetAttachRejectsForeignAndRevokedTokens(t *testing.T) {
	service, ctx := setupFleetService(t)
	fleet, err := service.CreateFleet(ctx, "org-1", "edge-nodes", "")
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := service.secrets.CreateMachineToken(ctx, "org-2", "", "", "foreign", []string{"tunnels:write"}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AttachToken(ctx, "org-1", fleet.ID, foreign.Token.ID, ""); err == nil {
		t.Fatal("foreign token was attached")
	}
	revoked, err := service.secrets.CreateMachineToken(ctx, "org-1", "", "", "revoked", []string{"tunnels:write"}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.secrets.RevokeMachineToken(ctx, "org-1", revoked.Token.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AttachToken(ctx, "org-1", fleet.ID, revoked.Token.ID, ""); err == nil {
		t.Fatal("revoked token was attached")
	}
}

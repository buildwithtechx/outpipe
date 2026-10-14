package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/workers"
)

func openFleetDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fleet.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.SecretMachineToken{}, &models.Tunnel{}, &models.MachineFleet{}, &models.MachineFleetToken{}, &models.MachineFleetTunnel{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestFleetSurvivesRelayRestartsAndCredentialRotations(t *testing.T) {
	db := openFleetDatabase(t)
	ctx := context.Background()
	secretRepo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := services.NewSecretService(secretRepo, "fleet-integration-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	tunnelRepo, err := repositories.NewTunnelRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tunnels, err := services.NewTunnelService(tunnelRepo)
	if err != nil {
		t.Fatal(err)
	}
	fleetRepo, err := repositories.NewMachineFleetRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	fleets, err := services.NewMachineFleetService(fleetRepo, secrets, tunnels)
	if err != nil {
		t.Fatal(err)
	}
	fleet, err := fleets.CreateFleet(ctx, "org-fleet", "hundred-nodes", "100-tunnel fleet")
	if err != nil {
		t.Fatal(err)
	}
	token, err := secrets.CreateMachineToken(ctx, "org-fleet", "", "", "fleet-token", []string{"tunnels:write"}, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fleets.AttachToken(ctx, "org-fleet", fleet.ID, token.Token.ID, "primary"); err != nil {
		t.Fatal(err)
	}
	specs := make([]services.FleetTunnelSpec, 0, 100)
	for i := 0; i < 100; i++ {
		specs = append(specs, services.FleetTunnelSpec{Name: fmt.Sprintf("node-%03d", i), Protocol: "http", TargetHost: "127.0.0.1", TargetPort: 3000 + i})
	}
	created, err := fleets.BulkCreateTunnels(ctx, "org-fleet", fleet.ID, "user-1", specs)
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 100 {
		t.Fatalf("created = %d", len(created))
	}

	// Simulate relay restarts: tunnels flap to disconnected, reconcile must
	// bring them back toward active without manual steps.
	for _, tunnel := range created {
		if err := tunnels.SetStatus(ctx, tunnel.ID, models.TunnelStatusDisconnected); err != nil {
			t.Fatal(err)
		}
	}
	result, err := fleets.Reconcile(ctx, "org-fleet", fleet.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Errors) > 0 {
		t.Fatalf("reconcile errors: %v", result.Errors)
	}
	if result.Reactivated != 100 {
		t.Fatalf("reactivated = %d: %+v", result.Reactivated, result)
	}

	// Simulate credential rotation: old token revoked, new token attached,
	// fleet keeps serving with zero manual tunnel steps.
	rotated, err := fleets.RotateFleetTokens(ctx, "org-fleet", fleet.ID, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rotated) != 1 {
		t.Fatalf("rotated = %d", len(rotated))
	}
	if _, err := secrets.VerifyMachineToken(ctx, token.Raw); err == nil {
		t.Fatal("revoked fleet token still verifies")
	}
	status, err := fleets.Status(ctx, "org-fleet", fleet.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Tokens) != 1 || status.Tokens[0].ID == token.Token.ID {
		t.Fatalf("fleet tokens = %+v", status.Tokens)
	}
	if len(status.LiveTunnels) != 100 {
		t.Fatalf("live tunnels = %d", len(status.LiveTunnels))
	}
	for _, drift := range status.Drift {
		t.Fatalf("unexpected drift after rotation: %+v", drift)
	}

	// The cron worker path reconciles every fleet without errors.
	job, err := workers.NewMachineFleetJob(fleets)
	if err != nil {
		t.Fatal(err)
	}
	if err := job.Run(ctx); err != nil {
		t.Fatalf("fleet cron job: %v", err)
	}
}

package services

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type captureRevocationRepository struct {
	*repositories.GormTunnelRepository
}

func (r *captureRevocationRepository) FindByID(ctx context.Context, id string) (models.Tunnel, error) {
	tunnel, err := r.GormTunnelRepository.FindByID(ctx, id)
	if err != nil {
		return models.Tunnel{}, err
	}
	if err := r.Revoke(ctx, id, time.Now().UTC()); err != nil {
		return models.Tunnel{}, err
	}
	return tunnel, nil
}

func TestCaptureSettingCannotUndoConcurrentRevocation(t *testing.T) {
	db := reviewTestDatabase(t, &models.Tunnel{})
	repo, err := repositories.NewTunnelRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tunnel := models.Tunnel{OrganizationID: "org", Name: "test", Protocol: models.TunnelProtocolHTTP, Status: models.TunnelStatusActive, TargetHost: "127.0.0.1", TargetPort: 3000, PublicHostname: "test.example.com"}
	if err := repo.Create(context.Background(), &tunnel); err != nil {
		t.Fatal(err)
	}
	svc, err := NewTunnelService(&captureRevocationRepository{repo})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetCapture(context.Background(), tunnel.ID, true); err == nil {
		t.Fatal("capture update accepted revoked tunnel")
	}
	stored, err := repo.FindByID(context.Background(), tunnel.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != models.TunnelStatusRevoked || stored.RevokedAt == nil || stored.CaptureEnabled {
		t.Fatal("capture update undid revocation")
	}
}

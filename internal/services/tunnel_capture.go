package services

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func (s *TunnelService) SetCapture(ctx context.Context, tunnelID string, enabled bool) (models.Tunnel, error) {
	tunnel, err := s.Find(ctx, tunnelID)
	if err != nil {
		return models.Tunnel{}, fmt.Errorf("find capture tunnel: %w", err)
	}
	repo, ok := s.tunnels.(interface {
		SetCapture(context.Context, string, bool) error
	})
	if !ok {
		return models.Tunnel{}, fmt.Errorf("repository does not support capture updates")
	}
	if tunnel.Status == models.TunnelStatusRevoked {
		return models.Tunnel{}, repositories.ErrNotFound
	}
	if err := repo.SetCapture(ctx, tunnelID, enabled); err != nil {
		return models.Tunnel{}, fmt.Errorf("update capture policy: %w", err)
	}
	tunnel.CaptureEnabled = enabled
	return tunnel, nil
}

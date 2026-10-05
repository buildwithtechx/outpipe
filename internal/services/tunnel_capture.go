package services

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
)

func (s *TunnelService) SetCapture(ctx context.Context, tunnelID string, enabled bool) (models.Tunnel, error) {
	tunnel, err := s.Find(ctx, tunnelID)
	if err != nil {
		return models.Tunnel{}, fmt.Errorf("find capture tunnel: %w", err)
	}
	if tunnel.Status == models.TunnelStatusRevoked {
		return models.Tunnel{}, fmt.Errorf("tunnel is revoked")
	}
	tunnel.CaptureEnabled = enabled
	if err := s.tunnels.Update(ctx, &tunnel); err != nil {
		return models.Tunnel{}, fmt.Errorf("update capture policy: %w", err)
	}
	return tunnel, nil
}

package services

import (
	"context"
	"fmt"
	"strings"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/security"
)

func hashTunnelPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		if password != "" {
			return "", fmt.Errorf("tunnel password cannot contain only whitespace")
		}
		return "", nil
	}
	if len(password) < 8 || len(password) > 256 {
		return "", fmt.Errorf("tunnel password must be between 8 and 256 characters")
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("hash tunnel password: %w", err)
	}
	return hash, nil
}

func (s *TunnelService) statusDeliveries(ctx context.Context, id string, status models.TunnelStatus) ([]models.WebhookDelivery, error) {
	var event models.WebhookEvent
	switch status {
	case models.TunnelStatusActive:
		event = models.WebhookEventTunnelConnected
	case models.TunnelStatusDisconnected:
		event = models.WebhookEventTunnelDisconnected
	default:
		return nil, nil
	}

	tunnel, err := s.tunnels.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find tunnel for webhook event: %w", err)
	}
	tunnel.Status = status

	if s.webhooks == nil {
		return nil, nil
	}

	deliveries, err := s.webhooks.prepareDeliveries(ctx, tunnel.OrganizationID, event, map[string]any{"tunnel": tunnel})
	if err != nil {
		return nil, fmt.Errorf("prepare webhook deliveries: %w", err)
	}
	return deliveries, nil
}

func validTunnelProtocol(protocol models.TunnelProtocol) bool {
	return protocol == models.TunnelProtocolHTTP || protocol == models.TunnelProtocolHTTPS || protocol == models.TunnelProtocolTCP || protocol == models.TunnelProtocolUDP
}

func validTunnelStatus(status models.TunnelStatus) bool {
	return status == models.TunnelStatusCreated || status == models.TunnelStatusConnecting || status == models.TunnelStatusActive || status == models.TunnelStatusDisconnected || status == models.TunnelStatusExpired || status == models.TunnelStatusRevoked
}

package services

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/validation"
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

func (s *TunnelService) GetCapturePolicy(ctx context.Context, tunnelID string) (models.CapturePolicy, error) {
	tunnel, err := s.Find(ctx, tunnelID)
	if err != nil {
		return models.CapturePolicy{}, fmt.Errorf("find capture tunnel: %w", err)
	}
	if s.capturePolicies == nil {
		return models.CapturePolicy{}, fmt.Errorf("capture policies are not configured")
	}
	policy, err := s.capturePolicies.FindByTunnel(ctx, tunnel.ID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return models.CapturePolicy{OrganizationID: tunnel.OrganizationID, TunnelID: tunnel.ID, Version: 0, Rules: "[]"}, nil
		}
		return models.CapturePolicy{}, err
	}
	return policy, nil
}

func (s *TunnelService) SetCapturePolicy(ctx context.Context, tunnelID, rules string) (models.CapturePolicy, error) {
	tunnel, err := s.Find(ctx, tunnelID)
	if err != nil {
		return models.CapturePolicy{}, fmt.Errorf("find capture tunnel: %w", err)
	}
	if tunnel.Status == models.TunnelStatusRevoked {
		return models.CapturePolicy{}, repositories.ErrNotFound
	}
	if s.capturePolicies == nil {
		return models.CapturePolicy{}, fmt.Errorf("capture policies are not configured")
	}
	parsed, err := validation.ParseCaptureRules(rules)
	if err != nil {
		return models.CapturePolicy{}, err
	}
	normalized := "[]"
	if len(parsed) > 0 {
		encoded, err := validation.MarshalCaptureRules(parsed)
		if err != nil {
			return models.CapturePolicy{}, err
		}
		normalized = encoded
	}
	policy := models.CapturePolicy{OrganizationID: tunnel.OrganizationID, TunnelID: tunnel.ID, Rules: normalized}
	if err := s.capturePolicies.Upsert(ctx, &policy); err != nil {
		return models.CapturePolicy{}, fmt.Errorf("save capture policy: %w", err)
	}
	return policy, nil
}

func (s *TunnelService) CaptureAllowed(ctx context.Context, tunnelID, method, path string) (bool, error) {
	tunnel, err := s.Find(ctx, tunnelID)
	if err != nil {
		return false, fmt.Errorf("find capture tunnel: %w", err)
	}
	if !tunnel.CaptureEnabled || tunnel.Status == models.TunnelStatusRevoked {
		return false, nil
	}
	if s.capturePolicies == nil {
		return true, nil
	}
	policy, err := s.capturePolicies.FindByTunnel(ctx, tunnel.ID)
	if err != nil {
		if err == repositories.ErrNotFound {
			return true, nil
		}
		return false, err
	}
	rules, err := validation.ParseCaptureRules(policy.Rules)
	if err != nil {
		return false, nil
	}
	return validation.CaptureAllowed(rules, method, path), nil
}

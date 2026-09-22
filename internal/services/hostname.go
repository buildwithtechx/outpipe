package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/repositories"
)

type HostnameAllocator struct {
	tunnels    repositories.TunnelRepository
	baseDomain string
}

func NewHostnameAllocator(tunnels repositories.TunnelRepository, baseDomain string) (*HostnameAllocator, error) {

	if tunnels == nil || strings.TrimSpace(baseDomain) == "" {
		return nil, fmt.Errorf("tunnel repository and base domain are required")
	}

	return &HostnameAllocator{tunnels: tunnels, baseDomain: strings.ToLower(strings.TrimSuffix(strings.TrimSpace(baseDomain), "."))}, nil
}

func (a *HostnameAllocator) Allocate(ctx context.Context, requested string) (string, error) {

	if requested != "" {

		if _, err := a.tunnels.FindByHostname(ctx, requested); err == repositories.ErrNotFound {
			return strings.ToLower(strings.TrimSuffix(requested, ".")), nil

		} else if err == nil {
			return "", fmt.Errorf("hostname is already allocated")
		} else {
			return "", fmt.Errorf("check requested hostname: %w", err)
		}
	}

	for attempt := 0; attempt < 5; attempt++ {
		candidate := uuid.NewString()[:12] + "." + a.baseDomain

		if _, err := a.tunnels.FindByHostname(ctx, candidate); err == repositories.ErrNotFound {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("allocate public hostname")
}

package services

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func (s *UptimeService) GetOwnedMonitor(ctx context.Context, orgID, id string) (*models.UptimeMonitor, error) {
	monitor, err := s.GetMonitor(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find organization monitor: %w", err)
	}
	if monitor.OrganizationID != orgID {
		return nil, repositories.ErrNotFound
	}
	return monitor, nil
}

func (s *UptimeService) AuthorizeIncident(ctx context.Context, orgID, id string) error {
	repo, ok := s.repo.(interface {
		IncidentOrganization(context.Context, string) (string, error)
	})
	if !ok {
		return fmt.Errorf("repository does not support incident authorization")
	}
	organizationID, err := repo.IncidentOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("find incident: %w", err)
	}
	if organizationID != orgID {
		return repositories.ErrNotFound
	}
	return nil
}

func (s *UptimeService) MonitorChecks(ctx context.Context, orgID, id string) ([]models.UptimeCheck, error) {
	if _, err := s.GetOwnedMonitor(ctx, orgID, id); err != nil {
		return nil, err
	}
	rows, err := s.repo.GetRecentChecks(ctx, id, 100)
	if err != nil {
		return nil, fmt.Errorf("list monitor checks: %w", err)
	}
	return rows, nil
}

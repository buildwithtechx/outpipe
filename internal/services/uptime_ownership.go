package services

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
)

func (s *UptimeService) GetOwnedMonitor(ctx context.Context, orgID, id string) (*models.UptimeMonitor, error) {
	monitor, err := s.GetMonitor(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find organization monitor: %w", err)
	}
	if monitor.OrganizationID != orgID {
		return nil, fmt.Errorf("monitor not found")
	}
	return monitor, nil
}

func (s *UptimeService) AuthorizeIncident(ctx context.Context, orgID, id string) error {
	incident, err := s.repo.GetIncident(ctx, id)
	if err != nil {
		return fmt.Errorf("find incident: %w", err)
	}
	if incident.OrganizationID != orgID {
		return fmt.Errorf("incident not found")
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

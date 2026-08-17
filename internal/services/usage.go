package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type UsageService struct {
	usage repositories.UsageRepository
}

var ErrUsageServiceRequired = fmt.Errorf("usage service is required")

func NewUsageService(usage repositories.UsageRepository) (*UsageService, error) {

	if usage == nil {
		return nil, ErrUsageServiceRequired
	}

	return &UsageService{usage: usage}, nil
}

func (s *UsageService) Record(ctx context.Context, event *models.UsageEvent) error {

	if event == nil || event.OrganizationID == "" || event.EventType == "" || event.Bytes < 0 || event.Connections < 0 {
		return fmt.Errorf("invalid usage event")
	}

	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	if err := s.usage.CreateEvent(ctx, event); err != nil {
		return fmt.Errorf("record usage event: %w", err)
	}

	return nil
}

func (s *UsageService) Snapshot(ctx context.Context, snapshot *models.UsageSnapshot) error {

	if snapshot == nil || snapshot.OrganizationID == "" || snapshot.PeriodStart.IsZero() || snapshot.PeriodEnd.IsZero() || snapshot.PeriodEnd.Before(snapshot.PeriodStart) {
		return fmt.Errorf("invalid usage snapshot")
	}

	if snapshot.TunnelCount < 0 || snapshot.ActiveConnections < 0 || snapshot.BandwidthBytes < 0 || snapshot.RequestCount < 0 || snapshot.ErrorCount < 0 {
		return fmt.Errorf("usage counters cannot be negative")
	}

	if err := s.usage.UpsertSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("save usage snapshot: %w", err)
	}

	return nil
}

func (s *UsageService) Aggregate(ctx context.Context, organizationID string, from, to time.Time) (models.UsageSnapshot, error) {

	if organizationID == "" || from.IsZero() || to.IsZero() || !to.After(from) {
		return models.UsageSnapshot{}, fmt.Errorf("organization and valid aggregation period are required")
	}

	snapshot, err := s.usage.AggregatePeriod(ctx, organizationID, from, to)

	if err != nil {
		return models.UsageSnapshot{}, fmt.Errorf("aggregate usage: %w", err)
	}

	if err := s.Snapshot(ctx, &snapshot); err != nil {
		return models.UsageSnapshot{}, err
	}

	return snapshot, nil
}

func (s *UsageService) ListEvents(ctx context.Context, organizationID string, from, to time.Time) ([]models.UsageEvent, error) {

	if organizationID == "" || from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("organization and valid usage period are required")
	}

	events, err := s.usage.ListEvents(ctx, organizationID, from, to)

	if err != nil {
		return nil, fmt.Errorf("list usage events: %w", err)
	}

	return events, nil
}

func (s *UsageService) FindSnapshot(ctx context.Context, organizationID string, periodStart time.Time) (models.UsageSnapshot, error) {
	snapshot, err := s.usage.FindSnapshot(ctx, organizationID, periodStart)

	if err != nil {
		return models.UsageSnapshot{}, fmt.Errorf("find usage snapshot: %w", err)
	}

	return snapshot, nil
}

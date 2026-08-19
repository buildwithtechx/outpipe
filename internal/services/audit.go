package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type AuditService struct {
	audit repositories.AuditRepository
}

func NewAuditService(audit repositories.AuditRepository) (*AuditService, error) {

	if audit == nil {
		return nil, fmt.Errorf("audit repository is required")
	}

	return &AuditService{audit: audit}, nil
}

func (s *AuditService) Record(ctx context.Context, event *models.AuditEvent) error {

	if event == nil || event.Action == "" || event.ResourceType == "" {
		return fmt.Errorf("audit action and resource type are required")
	}

	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	if err := s.audit.Create(ctx, event); err != nil {
		return fmt.Errorf("record audit event: %w", err)
	}

	return nil
}

func (s *AuditService) ListByOrganization(ctx context.Context, organizationID string, from, to time.Time, limit int) ([]models.AuditEvent, error) {

	if organizationID == "" || from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, fmt.Errorf("organization and valid audit period are required")
	}

	events, err := s.audit.ListByOrganization(ctx, organizationID, from, to, limit)

	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}

	return events, nil
}

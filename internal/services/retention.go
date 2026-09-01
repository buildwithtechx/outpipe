package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/repositories"
)

type RetentionService struct {
	organizations repositories.OrganizationRepository
	billing       repositories.BillingRepository
	usage         repositories.UsageRepository
	audit         repositories.AuditRepository
	now           func() time.Time
}

func NewRetentionService(organizations repositories.OrganizationRepository, billing repositories.BillingRepository, usage repositories.UsageRepository, audit repositories.AuditRepository) (*RetentionService, error) {

	if organizations == nil || billing == nil || usage == nil || audit == nil {
		return nil, fmt.Errorf("retention repositories are required")
	}

	return &RetentionService{organizations: organizations, billing: billing, usage: usage, audit: audit, now: time.Now}, nil
}

func (s *RetentionService) Enforce(ctx context.Context, now time.Time) error {
	organizations, err := s.organizations.List(ctx)

	if err != nil {
		return fmt.Errorf("list organizations for retention: %w", err)
	}

	for _, organization := range organizations {
		subscription, err := s.billing.FindSubscription(ctx, organization.ID)

		if err != nil {

			if err == repositories.ErrNotFound {
				continue
			}

			return fmt.Errorf("find retention subscription: %w", err)
		}

		plan, err := s.billing.FindPlanByID(ctx, subscription.PlanID)

		if err != nil {
			return fmt.Errorf("find retention plan: %w", err)
		}

		if plan.RetentionDays <= 0 {
			continue
		}

		cutoff := now.AddDate(0, 0, -plan.RetentionDays)

		if _, err := s.usage.DeleteBefore(ctx, organization.ID, cutoff); err != nil {
			return fmt.Errorf("enforce usage retention: %w", err)
		}

		if _, err := s.audit.DeleteBefore(ctx, organization.ID, cutoff); err != nil {
			return fmt.Errorf("enforce audit retention: %w", err)
		}
	}

	return nil
}

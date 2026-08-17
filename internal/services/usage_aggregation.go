package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/repositories"
)

type UsageAggregationService struct {
	organizations repositories.OrganizationRepository
	usage         *UsageService
	billing       repositories.BillingRepository
	alerts        *AlertService
	now           func() time.Time
}

func NewUsageAggregationService(organizations repositories.OrganizationRepository, usage *UsageService) (*UsageAggregationService, error) {

	if organizations == nil || usage == nil {
		return nil, fmt.Errorf("organization repository and usage service are required")
	}

	return &UsageAggregationService{organizations: organizations, usage: usage, now: time.Now}, nil
}

func (s *UsageAggregationService) SetAlerts(alerts *AlertService) { s.alerts = alerts }
func (s *UsageAggregationService) SetBilling(billing repositories.BillingRepository) {
	s.billing = billing
}

func (s *UsageAggregationService) Aggregate(ctx context.Context, now time.Time) error {
	periodEnd := now.UTC().Truncate(time.Hour)
	periodStart := periodEnd.Add(-time.Hour)
	organizations, err := s.organizations.List(ctx)

	if err != nil {
		return fmt.Errorf("list organizations for usage aggregation: %w", err)
	}

	for _, organization := range organizations {
		snapshot, err := s.usage.Aggregate(ctx, organization.ID, periodStart, periodEnd)

		if err != nil {
			return fmt.Errorf("aggregate organization %s: %w", organization.ID, err)
		}

		if s.billing != nil && s.alerts != nil {
			sub, err := s.billing.FindSubscription(ctx, organization.ID)

			if err == nil {
				plan, err := s.billing.FindPlan(ctx, sub.PlanID)

				if err == nil && plan.BandwidthBytes > 0 && snapshot.BandwidthBytes > plan.BandwidthBytes {
					_ = s.alerts.AlertQuotaInconsistency(ctx, organization.ID, snapshot.BandwidthBytes, plan.BandwidthBytes)
				}
			}
		}
	}

	return nil
}

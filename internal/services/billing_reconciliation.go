package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type SubscriptionReconcilerService struct {
	billing repositories.BillingRepository
	alerts  *AlertService
	now     func() time.Time
}

func NewSubscriptionReconcilerService(billing repositories.BillingRepository, alerts *AlertService) (*SubscriptionReconcilerService, error) {

	if billing == nil {
		return nil, fmt.Errorf("billing repository is required")
	}

	return &SubscriptionReconcilerService{billing: billing, alerts: alerts, now: time.Now}, nil
}

func (s *SubscriptionReconcilerService) Reconcile(ctx context.Context, _ time.Time) error {
	subscriptions, err := s.billing.ListSubscriptions(ctx)

	if err != nil {
		return fmt.Errorf("list subscriptions for reconciliation: %w", err)
	}

	for _, sub := range subscriptions {
		subCopy := sub

		if subCopy.Status == models.SubscriptionStatusPastDue || subCopy.Status == models.SubscriptionStatusCanceled {

			if s.alerts != nil {

				if err := s.alerts.AlertRepeatedPaymentFailure(ctx, subCopy.OrganizationID, subCopy.ID, 1); err != nil {
					return fmt.Errorf("alert payment failure: %w", err)
				}
			}
		}
	}

	return nil
}

package services

import (
	"context"
	"fmt"
	"strings"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func (s *BillingService) Entitlements(ctx context.Context, organizationID string) (models.Plan, models.Subscription, error) {
	subscription, err := s.billing.FindSubscription(ctx, organizationID)
	if err != nil {
		return models.Plan{}, models.Subscription{}, fmt.Errorf("find subscription: %w", err)
	}

	entitled := subscription.Status == models.SubscriptionStatusActive || subscription.Status == models.SubscriptionStatusTrialing
	if !entitled && subscription.CurrentPeriodEnd != nil {
		entitled = (subscription.Status == models.SubscriptionStatusPastDue || subscription.Status == models.SubscriptionStatusCanceled) && s.now().Before(subscription.CurrentPeriodEnd.Add(s.gracePeriod))
	}
	if !entitled {
		return models.Plan{}, models.Subscription{}, fmt.Errorf("subscription is not entitled")
	}

	plan, err := s.billing.FindPlanByID(ctx, subscription.PlanID)
	if err != nil {
		return models.Plan{}, models.Subscription{}, fmt.Errorf("find subscription plan: %w", err)
	}
	if !plan.Active {
		return models.Plan{}, models.Subscription{}, fmt.Errorf("subscription plan is inactive")
	}

	return plan, subscription, nil
}

func (s *BillingService) ProvisionFreeSubscription(ctx context.Context, organizationID string) error {
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return fmt.Errorf("organization id is required")
	}

	if _, err := s.billing.FindSubscription(ctx, organizationID); err == nil {
		return nil
	} else if err != repositories.ErrNotFound {
		return fmt.Errorf("find organization subscription: %w", err)
	}

	plan, err := s.ensureFreePlan(ctx)
	if err != nil {
		return err
	}

	subscription := &models.Subscription{
		OrganizationID:  organizationID,
		PlanID:          plan.ID,
		Provider:        models.BillingProviderInternal,
		ProviderSubID:   "free:" + organizationID,
		Status:          models.SubscriptionStatusActive,
		BillingInterval: models.BillingIntervalMonth,
	}
	if err := s.billing.CreateSubscription(ctx, subscription); err != nil {
		if _, retryErr := s.billing.FindSubscription(ctx, organizationID); retryErr == nil {
			return nil
		}
		return fmt.Errorf("create free subscription: %w", err)
	}

	return nil
}

func (s *BillingService) ensureFreePlan(ctx context.Context) (models.Plan, error) {
	if plan, err := s.billing.FindPlan(ctx, "free"); err == nil {
		return plan, nil
	} else if err != repositories.ErrNotFound {
		return models.Plan{}, fmt.Errorf("find free plan: %w", err)
	}

	plan := models.Plan{Key: "free", Name: "Free", Currency: "USD", MaxTunnels: 2, MaxDomains: 0, MaxMembers: 1, MaxConnections: 10, BandwidthBytes: 2 * 1024 * 1024 * 1024, RetentionDays: 3, Features: `{}`, Active: true}
	if err := s.billing.CreatePlan(ctx, &plan); err != nil {
		if current, retryErr := s.billing.FindPlan(ctx, "free"); retryErr == nil {
			return current, nil
		}
		return models.Plan{}, fmt.Errorf("create free plan: %w", err)
	}

	return plan, nil
}

func (s *BillingService) RecordEvent(ctx context.Context, event *models.BillingEvent) (bool, error) {
	if event == nil || event.Provider == "" || event.ProviderEventID == "" || event.PayloadHash == "" {
		return false, fmt.Errorf("complete billing event is required")
	}

	_, err := s.billing.FindBillingEvent(ctx, event.Provider, event.ProviderEventID)
	if err == nil {
		return false, nil
	}
	if err != repositories.ErrNotFound {
		return false, fmt.Errorf("check billing event: %w", err)
	}
	if err := s.billing.CreateBillingEvent(ctx, event); err != nil {
		return false, fmt.Errorf("record billing event: %w", err)
	}

	return true, nil
}

func (s *BillingService) MarkProcessed(ctx context.Context, eventID string) error {
	if err := s.billing.MarkBillingEventProcessed(ctx, eventID, s.now()); err != nil {
		return fmt.Errorf("mark billing event processed: %w", err)
	}
	return nil
}

package services

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/models"
)

type BillingGateway interface {
	Checkout(context.Context, models.Plan, string) (string, error)
	Portal(context.Context, models.BillingProvider, string) (string, error)
	Cancel(context.Context, models.BillingProvider, string) error
	Resume(context.Context, models.BillingProvider, string) error
}

func (s *BillingService) SetGateway(gateway BillingGateway) { s.gateway = gateway }

func (s *BillingService) Checkout(ctx context.Context, organizationID, planKey, billingInterval string) (string, error) {

	if s.gateway == nil {
		return "", fmt.Errorf("billing gateway is not configured")
	}

	billingInterval, err := normalizeBillingInterval(billingInterval)

	if err != nil {
		return "", err
	}

	plan, err := s.billing.FindPlan(ctx, planKey)

	if err != nil {
		return "", fmt.Errorf("find checkout plan: %w", err)
	}

	plan.BillingInterval = billingInterval
	url, err := s.gateway.Checkout(ctx, plan, organizationID)

	if err != nil {
		return "", fmt.Errorf("create checkout: %w", err)
	}

	return url, nil
}

func normalizeBillingInterval(interval string) (string, error) {

	if interval == "" {
		return models.BillingIntervalMonth, nil
	}

	if interval != models.BillingIntervalMonth && interval != models.BillingIntervalYear {
		return "", fmt.Errorf("unsupported billing interval %q (use %q or %q)", interval, models.BillingIntervalMonth, models.BillingIntervalYear)
	}

	return interval, nil
}

func (s *BillingService) Portal(ctx context.Context, organizationID string) (string, error) {

	if s.gateway == nil {
		return "", fmt.Errorf("billing gateway is not configured")
	}

	subscription, err := s.billing.FindSubscription(ctx, organizationID)

	if err != nil {
		return "", fmt.Errorf("find subscription for portal: %w", err)
	}

	customerID := subscription.ProviderCustomerID

	if s.billingSecrets != nil && customerID != "" {
		customerID, err = s.billingSecrets.Open(customerID)

		if err != nil {
			return "", fmt.Errorf("decrypt billing customer: %w", err)
		}
	}

	url, err := s.gateway.Portal(ctx, subscription.Provider, customerID)

	if err != nil {
		return "", fmt.Errorf("create billing portal: %w", err)
	}

	return url, nil
}

func (s *BillingService) Cancel(ctx context.Context, organizationID string) error {

	if s.gateway == nil {
		return fmt.Errorf("billing gateway is not configured")
	}

	subscription, err := s.billing.FindSubscription(ctx, organizationID)

	if err != nil {
		return fmt.Errorf("find subscription to cancel: %w", err)
	}

	return s.gateway.Cancel(ctx, subscription.Provider, subscription.ProviderSubID)
}

func (s *BillingService) Resume(ctx context.Context, organizationID string) error {

	if s.gateway == nil {
		return fmt.Errorf("billing gateway is not configured")
	}

	subscription, err := s.billing.FindSubscription(ctx, organizationID)

	if err != nil {
		return fmt.Errorf("find subscription to resume: %w", err)
	}

	return s.gateway.Resume(ctx, subscription.Provider, subscription.ProviderSubID)
}

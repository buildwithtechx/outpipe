package billing

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/models"
)

type GatewayConfig struct {
	Polar    *PolarClient
	Paystack *PaystackClient
	Email    func(context.Context, string) (string, error)
}

type Gateway struct {
	polar    *PolarClient
	paystack *PaystackClient
	email    func(context.Context, string) (string, error)
}

func NewGateway(cfg GatewayConfig) (*Gateway, error) {
	if cfg.Polar == nil && cfg.Paystack == nil {
		return nil, fmt.Errorf("at least one billing provider is required")
	}
	return &Gateway{polar: cfg.Polar, paystack: cfg.Paystack, email: cfg.Email}, nil
}

func (g *Gateway) Checkout(ctx context.Context, plan models.Plan, organizationID string) (string, error) {
	if plan.Currency == "NGN" {
		if g.paystack == nil || g.email == nil {
			return "", fmt.Errorf("paystack checkout is not configured")
		}
		email, err := g.email(ctx, organizationID)
		if err != nil {
			return "", fmt.Errorf("resolve billing email: %w", err)
		}
		transaction, err := g.paystack.InitializeTransaction(ctx, email, plan.PriceMinor, map[string]any{"organization_id": organizationID, "plan_key": plan.Key})
		if err != nil {
			return "", err
		}
		return transaction.AuthorizationURL, nil
	}
	if g.polar == nil {
		return "", fmt.Errorf("polar checkout is not configured")
	}
	return g.polar.Checkout(ctx, plan, organizationID)
}

func (g *Gateway) Portal(ctx context.Context, provider models.BillingProvider, customerID string) (string, error) {
	switch provider {
	case models.BillingProviderPolar:
		if g.polar == nil {
			return "", fmt.Errorf("polar is not configured")
		}
		return g.polar.Portal(ctx, customerID)
	case models.BillingProviderPaystack:
		if g.paystack == nil {
			return "", fmt.Errorf("paystack is not configured")
		}
		return g.paystack.Portal(ctx, customerID)
	default:
		return "", fmt.Errorf("unsupported billing provider %q", provider)
	}
}

func (g *Gateway) Cancel(ctx context.Context, provider models.BillingProvider, subscriptionID string) error {
	switch provider {
	case models.BillingProviderPolar:
		if g.polar == nil {
			return fmt.Errorf("polar is not configured")
		}
		return g.polar.Cancel(ctx, subscriptionID)
	case models.BillingProviderPaystack:
		if g.paystack == nil {
			return fmt.Errorf("paystack is not configured")
		}
		return g.paystack.Cancel(ctx, subscriptionID)
	default:
		return fmt.Errorf("unsupported billing provider %q", provider)
	}
}

func (g *Gateway) Resume(ctx context.Context, provider models.BillingProvider, subscriptionID string) error {
	switch provider {
	case models.BillingProviderPolar:
		if g.polar == nil {
			return fmt.Errorf("polar is not configured")
		}
		return g.polar.Resume(ctx, subscriptionID)
	case models.BillingProviderPaystack:
		if g.paystack == nil {
			return fmt.Errorf("paystack is not configured")
		}
		return g.paystack.Resume(ctx, subscriptionID)
	default:
		return fmt.Errorf("unsupported billing provider %q", provider)
	}
}

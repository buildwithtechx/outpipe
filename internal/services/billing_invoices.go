package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"outpipe.dev/outpipe/internal/models"
)

func (s *BillingService) syncInvoice(ctx context.Context, event *models.BillingEvent, subscription *models.Subscription, transition *BillingTransition) {
	if subscription == nil || transition == nil || transition.ProviderInvoice == "" {
		return
	}

	eventType := strings.ToLower(event.EventType)
	if !strings.Contains(eventType, "paid") && !strings.Contains(eventType, "success") && !strings.Contains(eventType, "order") {
		return
	}

	paidAt := s.now()
	if transition.PaidAt != nil {
		paidAt = *transition.PaidAt
	}

	currency := transition.Currency
	if currency == "" {
		currency = "USD"
	}

	invoice := &models.Invoice{
		OrganizationID:  subscription.OrganizationID,
		SubscriptionID:  subscription.ID,
		Provider:        transition.Provider,
		ProviderInvoice: transition.ProviderInvoice,
		AmountMinor:     transition.AmountMinor,
		Currency:        currency,
		Status:          "paid",
		InvoiceURL:      transition.InvoiceURL,
		PaidAt:          &paidAt,
	}
	if err := s.billing.CreateInvoice(ctx, invoice); err != nil {
		slog.Default().WarnContext(ctx, "create invoice failed", "organization_id", subscription.OrganizationID, "error", err)
	}
}

func (s *BillingService) ListInvoices(ctx context.Context, organizationID string) ([]models.Invoice, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization id is required")
	}

	invoices, err := s.billing.ListInvoicesByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	return invoices, nil
}

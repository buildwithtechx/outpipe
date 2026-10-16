package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/pkg/utils"
)

func (s *BillingService) notifyTransition(ctx context.Context, event *models.BillingEvent, subscription *models.Subscription, transition *BillingTransition, previousPlan string) {
	if s.mailer == nil || subscription == nil {
		return
	}

	if event == nil || s.notifications == nil {
		if err := s.mailer.SendBillingUpdate(ctx, subscription.OrganizationID, string(subscription.Status)); err != nil {
			slog.Default().WarnContext(ctx, "queue billing email failed", "organization_id", subscription.OrganizationID, "error", err)
		}
		return
	}

	target, err := s.notifications(ctx, subscription.OrganizationID)
	if err != nil {
		slog.Default().WarnContext(ctx, "resolve billing email recipient failed", "organization_id", subscription.OrganizationID, "error", err)
		return
	}

	eventType := strings.ToLower(event.EventType)

	if (transition != nil && transition.Status == models.SubscriptionStatusPastDue) || strings.Contains(eventType, "fail") || strings.Contains(eventType, "past_due") {
		planName := "Subscription"
		amount := utils.FormatMinorAmount(0, "USD")

		if subscription.PlanID != "" {
			if plan, planErr := s.billing.FindPlanByID(ctx, subscription.PlanID); planErr == nil {
				planName = plan.Name
				amount = utils.FormatMinorAmount(plan.PriceMinor, plan.Currency)
			}
		}

		attempts := 0
		attemptsKnown := false

		if transition != nil {
			attempts = transition.AttemptsRemaining
			attemptsKnown = transition.AttemptsKnown

			if transition.AmountMinor > 0 && transition.Currency != "" {
				amount = utils.FormatMinorAmount(transition.AmountMinor, transition.Currency)
			}
		}

		if err := s.mailer.SendPaymentFailed(ctx, target.Email, target.Name, planName, amount, target.BillingURL, attempts, attemptsKnown); err != nil {
			slog.Default().WarnContext(ctx, "queue payment failure email failed", "organization_id", subscription.OrganizationID, "error", err)
		}
		return
	}

	if strings.Contains(eventType, "downgrade") || strings.Contains(eventType, "reset") || strings.Contains(eventType, "revoke") {
		if transition != nil && transition.PreviousPlan != "" {
			previousPlan = transition.PreviousPlan
		}

		if err := s.mailer.SendSubscriptionReset(ctx, target.Email, target.Name, target.OrganizationName, previousPlan, s.dashboardURL); err != nil {
			slog.Default().WarnContext(ctx, "queue subscription email failed", "organization_id", subscription.OrganizationID, "error", err)
		}
		return
	}

	if transition != nil && transition.ProviderInvoice != "" && (strings.Contains(eventType, "paid") || strings.Contains(eventType, "success") || strings.Contains(eventType, "order")) {
		amount := utils.FormatMinorAmount(transition.AmountMinor, transition.Currency)
		if transition.AmountMinor == 0 {
			if plan, planErr := s.billing.FindPlanByID(ctx, subscription.PlanID); planErr == nil {
				amount = utils.FormatMinorAmount(plan.PriceMinor, plan.Currency)
			}
		}
		if err := s.mailer.SendInvoiceReceipt(ctx, target.Email, target.Name, target.OrganizationName, amount, transition.InvoiceURL); err != nil {
			return
		}
	}

	if err := s.mailer.SendBillingUpdate(ctx, subscription.OrganizationID, string(subscription.Status)); err != nil {
		slog.Default().WarnContext(ctx, "queue billing email failed", "organization_id", subscription.OrganizationID, "error", err)
	}
}

func (s *BillingService) applyTransition(subscription *models.Subscription, transition BillingTransition) error {
	subscription.Status = transition.Status

	if transition.ProviderCustomer != "" {
		if s.billingSecrets == nil {
			return fmt.Errorf("billing secret protector is not configured")
		}

		encrypted, err := s.billingSecrets.Seal(transition.ProviderCustomer)
		if err != nil {
			return fmt.Errorf("encrypt provider customer: %w", err)
		}

		subscription.ProviderCustomerID = encrypted
	}

	subscription.ProviderProductID = transition.ProviderProduct
	subscription.CurrentPeriodEnd = transition.CurrentPeriodEnd
	subscription.CancelAtPeriodEnd = transition.CancelAtPeriodEnd

	if transition.Status == models.SubscriptionStatusCanceled || transition.Status == models.SubscriptionStatusExpired {
		now := s.now()
		subscription.CanceledAt = &now
	}

	if transition.ProviderAuthorization != "" {
		if s.secrets == nil {
			return fmt.Errorf("billing secret protector is not configured")
		}

		encrypted, err := s.secrets.Seal(transition.ProviderAuthorization)
		if err != nil {
			return fmt.Errorf("encrypt provider authorization: %w", err)
		}

		subscription.ProviderAuthCode = encrypted
	}

	if transition.BillingInterval != "" {
		subscription.BillingInterval = transition.BillingInterval
	}

	return nil
}

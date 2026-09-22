package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/infra/telemetry"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/pkg/utils"
)

type BillingService struct {
	billing        repositories.BillingRepository
	gateway        BillingGateway
	now            func() time.Time
	secrets        SecretProtector
	billingSecrets billingSecretProtector
	gracePeriod    time.Duration
	mailer         BillingMailer
	notifications  BillingNotificationResolver
	dashboardURL   string
	metrics        *telemetry.MetricsExporter
}

func (s *BillingService) ListPlans(ctx context.Context) ([]models.Plan, error) {
	plans, err := s.billing.ListActivePlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("list billing plans: %w", err)
	}
	return plans, nil
}

type BillingMailer interface {
	SendBillingUpdate(context.Context, string, string) error
	SendPaymentFailed(context.Context, string, string, string, string, string, int, bool) error
	SendSubscriptionReset(context.Context, string, string, string, string, string) error
	SendInvoiceReceipt(context.Context, string, string, string, string, string) error
}

type BillingNotificationTarget struct {
	Email            string
	Name             string
	OrganizationName string
	BillingURL       string
}

type BillingNotificationResolver func(context.Context, string) (BillingNotificationTarget, error)

type billingSecretProtector interface {
	Seal(string) (string, error)
	Open(string) (string, error)
}

type BillingTransition struct {
	Provider              models.BillingProvider
	ProviderSubscription  string
	ProviderCustomer      string
	ProviderProduct       string
	ProviderInvoice       string
	InvoiceURL            string
	Status                models.SubscriptionStatus
	CurrentPeriodEnd      *time.Time
	CancelAtPeriodEnd     bool
	ProviderAuthorization string
	EventType             string
	AmountMinor           int64
	Currency              string
	AttemptsRemaining     int
	AttemptsKnown         bool
	PreviousPlan          string
	PaidAt                *time.Time
	BillingInterval       string
}

func (s *BillingService) ProcessWebhook(ctx context.Context, event *models.BillingEvent, transition *BillingTransition) (bool, error) {

	if event == nil || event.Provider == "" || event.ProviderEventID == "" {
		return false, fmt.Errorf("complete billing event is required")
	}
	if s.metrics != nil {
		s.metrics.IncCounter("outpipe_billing_events_received_total", 1)
	}

	if existing, err := s.billing.FindBillingEvent(ctx, event.Provider, event.ProviderEventID); err == nil {
		if existing.ProcessedAt == nil && existing.FailureReason != "" {
			// Failed events are intentionally retried; the unique key still
			// prevents concurrent successful processing from being duplicated.
		} else {
			if s.metrics != nil {
				s.metrics.IncCounter("outpipe_billing_events_duplicate_total", 1)
			}
			return false, nil
		}

	} else if err != repositories.ErrNotFound {
		return false, fmt.Errorf("check billing event: %w", err)
	}

	var subscription *models.Subscription
	var previousPlan string

	if transition != nil && transition.ProviderSubscription != "" && transition.Status != "" {
		current, err := s.billing.FindSubscriptionByProvider(ctx, transition.Provider, transition.ProviderSubscription)

		if err != nil {
			s.recordWebhookFailure(ctx, event, err)
			return false, fmt.Errorf("find subscription transition: %w", err)
		}

		if plan, planErr := s.billing.FindPlanByID(ctx, current.PlanID); planErr == nil {
			previousPlan = plan.Name
		}

		if err := s.applyTransition(&current, *transition); err != nil {
			s.recordWebhookFailure(ctx, event, err)
			return false, err
		}

		subscription = &current
	}

	if err := s.billing.ApplyBillingEvent(ctx, event, subscription); err != nil {
		if errors.Is(err, repositories.ErrBillingEventDuplicate) {
			if s.metrics != nil {
				s.metrics.IncCounter("outpipe_billing_events_duplicate_total", 1)
			}
			return false, nil
		}
		if s.metrics != nil {
			s.metrics.IncCounter("outpipe_billing_events_failed_total", 1)
		}
		s.recordWebhookFailure(ctx, event, err)
		return false, fmt.Errorf("apply billing webhook transaction: %w", err)
	}
	if s.metrics != nil {
		s.metrics.IncCounter("outpipe_billing_events_processed_total", 1)
	}

	if subscription != nil {
		s.notifyTransition(ctx, event, subscription, transition, previousPlan)
		s.syncInvoice(ctx, event, subscription, transition)
	}

	return true, nil
}

func (s *BillingService) recordWebhookFailure(ctx context.Context, event *models.BillingEvent, err error) {
	if event == nil || err == nil {
		return
	}
	if failureErr := s.billing.MarkBillingEventFailed(ctx, event.Provider, event.ProviderEventID, err.Error()); failureErr != nil {
		slog.Default().WarnContext(ctx, "record billing webhook failure failed", "provider", event.Provider, "event_id", event.ProviderEventID, "error", failureErr)
	}
}

func NewBillingService(billing repositories.BillingRepository) (*BillingService, error) {

	if billing == nil {
		return nil, fmt.Errorf("billing repository is required")
	}

	return &BillingService{billing: billing, now: time.Now, gracePeriod: 72 * time.Hour}, nil
}

func (s *BillingService) SetSecretProtector(protector SecretProtector) { s.secrets = protector }
func (s *BillingService) SetBillingSecretProtector(protector billingSecretProtector) {
	s.billingSecrets = protector
}
func (s *BillingService) SetGracePeriod(grace time.Duration) {

	if grace >= 0 {
		s.gracePeriod = grace
	}
}
func (s *BillingService) SetMailer(mailer BillingMailer) { s.mailer = mailer }
func (s *BillingService) SetNotificationResolver(resolver BillingNotificationResolver, dashboardURL string) {
	s.notifications = resolver
	s.dashboardURL = strings.TrimRight(dashboardURL, "/")
}

func (s *BillingService) SetMetrics(metrics *telemetry.MetricsExporter) { s.metrics = metrics }

func (s *BillingService) ApplyTransition(ctx context.Context, transition BillingTransition) error {

	if transition.Provider == "" || transition.ProviderSubscription == "" || transition.Status == "" {
		return fmt.Errorf("complete billing transition is required")
	}

	subscription, err := s.billing.FindSubscriptionByProvider(ctx, transition.Provider, transition.ProviderSubscription)

	if err != nil {
		return fmt.Errorf("find subscription transition: %w", err)
	}

	if err := s.applyTransition(&subscription, transition); err != nil {
		return err
	}

	if err := s.billing.SaveSubscription(ctx, &subscription); err != nil {
		return fmt.Errorf("save billing transition: %w", err)
	}

	previousPlan := ""

	if plan, planErr := s.billing.FindPlanByID(ctx, subscription.PlanID); planErr == nil {
		previousPlan = plan.Name
	}

	s.notifyTransition(ctx, &models.BillingEvent{EventType: transition.EventType}, &subscription, &transition, previousPlan)
	return nil
}

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

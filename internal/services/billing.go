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
	CancelAtPeriodEnd     *bool
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

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

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
}

type BillingMailer interface {
	SendBillingUpdate(context.Context, string, string) error
	SendPaymentFailed(context.Context, string, string, string, string, string, int, bool) error
	SendSubscriptionReset(context.Context, string, string, string, string, string) error
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

	if _, err := s.billing.FindBillingEvent(ctx, event.Provider, event.ProviderEventID); err == nil {
		return false, nil

	} else if err != repositories.ErrNotFound {
		return false, fmt.Errorf("check billing event: %w", err)
	}

	var subscription *models.Subscription
	var previousPlan string

	if transition != nil && transition.ProviderSubscription != "" && transition.Status != "" {
		current, err := s.billing.FindSubscriptionByProvider(ctx, transition.Provider, transition.ProviderSubscription)

		if err != nil {
			return false, fmt.Errorf("find subscription transition: %w", err)
		}

		if plan, planErr := s.billing.FindPlanByID(ctx, current.PlanID); planErr == nil {
			previousPlan = plan.Name
		}

		if err := s.applyTransition(&current, *transition); err != nil {
			return false, err
		}

		subscription = &current
	}

	if err := s.billing.ApplyBillingEvent(ctx, event, subscription); err != nil {
		return false, fmt.Errorf("apply billing webhook transaction: %w", err)
	}

	if subscription != nil {
		s.notifyTransition(ctx, event, subscription, transition, previousPlan)
		s.syncInvoice(ctx, event, subscription, transition)
	}

	return true, nil
}

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

	invoice := &models.Invoice{OrganizationID: subscription.OrganizationID, SubscriptionID: subscription.ID, Provider: transition.Provider, ProviderInvoice: transition.ProviderInvoice, AmountMinor: transition.AmountMinor, Currency: currency, Status: "paid", InvoiceURL: transition.InvoiceURL, PaidAt: &paidAt}
	_ = s.billing.CreateInvoice(ctx, invoice)
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
		_ = s.mailer.SendBillingUpdate(ctx, subscription.OrganizationID, string(subscription.Status))
		return
	}

	target, err := s.notifications(ctx, subscription.OrganizationID)

	if err != nil {
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

		_ = s.mailer.SendPaymentFailed(ctx, target.Email, target.Name, planName, amount, target.BillingURL, attempts, attemptsKnown)
		return
	}

	if strings.Contains(eventType, "downgrade") || strings.Contains(eventType, "reset") || strings.Contains(eventType, "revoke") {

		if transition != nil && transition.PreviousPlan != "" {
			previousPlan = transition.PreviousPlan
		}

		_ = s.mailer.SendSubscriptionReset(ctx, target.Email, target.Name, target.OrganizationName, previousPlan, s.dashboardURL)
		return
	}

	_ = s.mailer.SendBillingUpdate(ctx, subscription.OrganizationID, string(subscription.Status))
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

	plan := models.Plan{
		Key:            "free",
		Name:           "Free",
		Currency:       "USD",
		MaxTunnels:     2,
		MaxDomains:     0,
		MaxMembers:     1,
		MaxConnections: 10,
		BandwidthBytes: 2 * 1024 * 1024 * 1024,
		RetentionDays:  3,
		Features:       `{}`,
		Active:         true,
	}

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

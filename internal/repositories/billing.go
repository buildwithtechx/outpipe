package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

var ErrBillingEventDuplicate = errors.New("billing event already recorded")

type BillingRepository interface {
	CreatePlan(context.Context, *models.Plan) error
	FindPlan(context.Context, string) (models.Plan, error)
	FindPlanByID(context.Context, string) (models.Plan, error)
	ListActivePlans(context.Context) ([]models.Plan, error)
	CreateSubscription(context.Context, *models.Subscription) error
	FindSubscription(context.Context, string) (models.Subscription, error)
	FindSubscriptionByProvider(context.Context, models.BillingProvider, string) (models.Subscription, error)
	ListSubscriptions(context.Context) ([]models.Subscription, error)
	SaveSubscription(context.Context, *models.Subscription) error
	FindBillingEvent(context.Context, models.BillingProvider, string) (models.BillingEvent, error)
	CreateBillingEvent(context.Context, *models.BillingEvent) error
	MarkBillingEventProcessed(context.Context, string, time.Time) error
	MarkBillingEventFailed(context.Context, models.BillingProvider, string, string) error
	ApplyBillingEvent(context.Context, *models.BillingEvent, *models.Subscription) error
	SaveCredential(context.Context, *models.BillingCredential) error
	FindCredential(context.Context, string, models.BillingProvider, string) (models.BillingCredential, error)
	CreateInvoice(context.Context, *models.Invoice) error
	ListInvoicesByOrganization(context.Context, string) ([]models.Invoice, error)
}

type GormBillingRepository struct{ db *gorm.DB }

func NewBillingRepository(db *gorm.DB) (*GormBillingRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormBillingRepository{db: db}, nil
}

func (r *GormBillingRepository) CreatePlan(ctx context.Context, plan *models.Plan) error {

	if plan == nil {
		return fmt.Errorf("plan is required")
	}

	return wrap(r.db.WithContext(ctx).Create(plan).Error, "create plan")
}

func (r *GormBillingRepository) FindPlan(ctx context.Context, key string) (models.Plan, error) {
	var plan models.Plan

	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&plan).Error; err != nil {
		return models.Plan{}, mapError(err)
	}

	return plan, nil
}

func (r *GormBillingRepository) FindPlanByID(ctx context.Context, id string) (models.Plan, error) {
	var plan models.Plan

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error; err != nil {
		return models.Plan{}, mapError(err)
	}

	return plan, nil
}

func (r *GormBillingRepository) ListActivePlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan

	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("price_minor ASC").Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("list active plans: %w", err)
	}

	return plans, nil
}

func (r *GormBillingRepository) CreateSubscription(ctx context.Context, subscription *models.Subscription) error {

	if subscription == nil {
		return fmt.Errorf("subscription is required")
	}

	return wrap(r.db.WithContext(ctx).Create(subscription).Error, "create subscription")
}

func (r *GormBillingRepository) FindSubscription(ctx context.Context, organizationID string) (models.Subscription, error) {
	var subscription models.Subscription

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).First(&subscription).Error; err != nil {
		return models.Subscription{}, mapError(err)
	}

	return subscription, nil
}

func (r *GormBillingRepository) ListSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	var subscriptions []models.Subscription

	if err := r.db.WithContext(ctx).Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *GormBillingRepository) FindSubscriptionByProvider(ctx context.Context, provider models.BillingProvider, providerID string) (models.Subscription, error) {
	var subscription models.Subscription

	if err := r.db.WithContext(ctx).Where("provider = ? AND provider_sub_id = ?", provider, providerID).First(&subscription).Error; err != nil {
		return models.Subscription{}, mapError(err)
	}

	return subscription, nil
}

func (r *GormBillingRepository) SaveSubscription(ctx context.Context, subscription *models.Subscription) error {

	if subscription == nil {
		return fmt.Errorf("subscription is required")
	}

	return wrap(r.db.WithContext(ctx).Save(subscription).Error, "save subscription")
}

func (r *GormBillingRepository) FindBillingEvent(ctx context.Context, provider models.BillingProvider, eventID string) (models.BillingEvent, error) {
	var event models.BillingEvent

	if err := r.db.WithContext(ctx).Where("provider = ? AND provider_event_id = ?", provider, eventID).First(&event).Error; err != nil {
		return models.BillingEvent{}, mapError(err)
	}

	return event, nil
}

func (r *GormBillingRepository) CreateBillingEvent(ctx context.Context, event *models.BillingEvent) error {

	if event == nil {
		return fmt.Errorf("billing event is required")
	}

	err := r.db.WithContext(ctx).Create(event).Error
	if isBillingEventDuplicate(err) {
		return ErrBillingEventDuplicate
	}
	return wrap(err, "create billing event")
}

func (r *GormBillingRepository) MarkBillingEventProcessed(ctx context.Context, id string, at time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.BillingEvent{}).Where("id = ?", id).Update("processed_at", at)

	if result.Error != nil {
		return fmt.Errorf("mark billing event processed: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormBillingRepository) MarkBillingEventFailed(ctx context.Context, provider models.BillingProvider, eventID, reason string) error {
	reason = strings.TrimSpace(reason)
	if len(reason) > 500 {
		reason = reason[:500]
	}

	result := r.db.WithContext(ctx).Model(&models.BillingEvent{}).
		Where("provider = ? AND provider_event_id = ?", provider, eventID).
		Updates(map[string]any{"failure_reason": reason, "processed_at": nil})
	if result.Error != nil {
		return fmt.Errorf("update billing event failure: %w", result.Error)
	}
	if result.RowsAffected == 1 {
		return nil
	}

	event := &models.BillingEvent{Provider: provider, ProviderEventID: eventID, EventType: "unknown", PayloadHash: "unknown", FailureReason: reason}
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil && !isBillingEventDuplicate(err) {
		return fmt.Errorf("create failed billing event: %w", err)
	}
	if err := r.db.WithContext(ctx).Model(&models.BillingEvent{}).
		Where("provider = ? AND provider_event_id = ?", provider, eventID).
		Updates(map[string]any{"failure_reason": reason, "processed_at": nil}).Error; err != nil {
		return fmt.Errorf("persist billing event failure: %w", err)
	}
	return nil
}

func (r *GormBillingRepository) ApplyBillingEvent(ctx context.Context, event *models.BillingEvent, subscription *models.Subscription) error {

	if event == nil {
		return fmt.Errorf("billing event is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(event).Error; err != nil {
			if isBillingEventDuplicate(err) {
				return ErrBillingEventDuplicate
			}
			return fmt.Errorf("create billing event: %w", err)
		}

		if subscription != nil {

			if err := tx.Save(subscription).Error; err != nil {
				return fmt.Errorf("save subscription transition: %w", err)
			}
		}

		at := time.Now().UTC()

		if err := tx.Model(&models.BillingEvent{}).Where("id = ?", event.ID).Update("processed_at", at).Error; err != nil {
			return fmt.Errorf("mark billing event processed: %w", err)
		}

		return nil
	})
}

func isBillingEventDuplicate(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "provider_event") ||
		strings.Contains(message, "billing_events.provider")
}

func (r *GormBillingRepository) SaveCredential(ctx context.Context, credential *models.BillingCredential) error {

	if credential == nil {
		return fmt.Errorf("billing credential is required")
	}

	return wrap(r.db.WithContext(ctx).Save(credential).Error, "save billing credential")
}

func (r *GormBillingRepository) FindCredential(ctx context.Context, organizationID string, provider models.BillingProvider, kind string) (models.BillingCredential, error) {
	var credential models.BillingCredential

	if err := r.db.WithContext(ctx).Where("organization_id = ? AND provider = ? AND kind = ? AND revoked_at IS NULL", organizationID, provider, kind).First(&credential).Error; err != nil {
		return models.BillingCredential{}, mapError(err)
	}

	return credential, nil
}

func (r *GormBillingRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {

	if invoice == nil {
		return fmt.Errorf("invoice is required")
	}

	return wrap(r.db.WithContext(ctx).Create(invoice).Error, "create invoice")
}

func (r *GormBillingRepository) ListInvoicesByOrganization(ctx context.Context, organizationID string) ([]models.Invoice, error) {
	var invoices []models.Invoice

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("list organization invoices: %w", err)
	}

	return invoices, nil
}

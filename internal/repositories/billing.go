package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type BillingRepository interface {
	CreatePlan(context.Context, *models.Plan) error
	FindPlan(context.Context, string) (models.Plan, error)
	ListActivePlans(context.Context) ([]models.Plan, error)
	FindSubscription(context.Context, string) (models.Subscription, error)
	FindSubscriptionByProvider(context.Context, models.BillingProvider, string) (models.Subscription, error)
	ListSubscriptions(context.Context) ([]models.Subscription, error)
	SaveSubscription(context.Context, *models.Subscription) error
	FindBillingEvent(context.Context, models.BillingProvider, string) (models.BillingEvent, error)
	CreateBillingEvent(context.Context, *models.BillingEvent) error
	MarkBillingEventProcessed(context.Context, string, time.Time) error
	ApplyBillingEvent(context.Context, *models.BillingEvent, *models.Subscription) error
	SaveCredential(context.Context, *models.BillingCredential) error
	FindCredential(context.Context, string, models.BillingProvider, string) (models.BillingCredential, error)
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

func (r *GormBillingRepository) ListActivePlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("price_minor ASC").Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("list active plans: %w", err)
	}
	return plans, nil
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
	return wrap(r.db.WithContext(ctx).Create(event).Error, "create billing event")
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

func (r *GormBillingRepository) ApplyBillingEvent(ctx context.Context, event *models.BillingEvent, subscription *models.Subscription) error {
	if event == nil {
		return fmt.Errorf("billing event is required")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
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

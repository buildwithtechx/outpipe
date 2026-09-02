package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

type WebhookRepository interface {
	Create(context.Context, *models.WebhookSubscription) error
	Update(context.Context, *models.WebhookSubscription) error
	FindByID(context.Context, string) (models.WebhookSubscription, error)
	ListByOrganization(context.Context, string) ([]models.WebhookSubscription, error)
	Delete(context.Context, string) error
	CreateDelivery(context.Context, *models.WebhookDelivery) error
	UpdateDelivery(context.Context, *models.WebhookDelivery) error
	ClaimPendingDeliveries(context.Context, time.Time, int) ([]models.WebhookDelivery, error)
	CountQueuedDeliveries(context.Context) (int64, error)
	ListDeliveries(context.Context, string) ([]models.WebhookDelivery, error)
}

func (r *GormWebhookRepository) CountQueuedDeliveries(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.WebhookDelivery{}).
		Where("status IN ?", []models.WebhookDeliveryStatus{models.WebhookDeliveryPending, models.WebhookDeliveryFailed, models.WebhookDeliverySending}).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count queued webhook deliveries: %w", err)
	}
	return count, nil
}

type GormWebhookRepository struct{ db *gorm.DB }

func NewWebhookRepository(db *gorm.DB) (*GormWebhookRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormWebhookRepository{db: db}, nil
}

func (r *GormWebhookRepository) Create(ctx context.Context, subscription *models.WebhookSubscription) error {

	if subscription == nil {
		return fmt.Errorf("webhook subscription is required")
	}

	return wrap(r.db.WithContext(ctx).Create(subscription).Error, "create webhook subscription")
}

func (r *GormWebhookRepository) Update(ctx context.Context, subscription *models.WebhookSubscription) error {
	if subscription == nil {
		return fmt.Errorf("webhook subscription is required")
	}

	return wrap(r.db.WithContext(ctx).Save(subscription).Error, "update webhook subscription")
}

func (r *GormWebhookRepository) FindByID(ctx context.Context, id string) (models.WebhookSubscription, error) {
	var subscription models.WebhookSubscription

	if err := r.db.WithContext(ctx).First(&subscription, "id = ?", id).Error; err != nil {
		return models.WebhookSubscription{}, mapError(err)
	}

	return subscription, nil
}

func (r *GormWebhookRepository) ListByOrganization(ctx context.Context, organizationID string) ([]models.WebhookSubscription, error) {
	var subscriptions []models.WebhookSubscription

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at DESC").Limit(DefaultListLimit).Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("list webhook subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *GormWebhookRepository) Delete(ctx context.Context, id string) error {

	if err := r.db.WithContext(ctx).Where("subscription_id = ?", id).Delete(&models.WebhookDelivery{}).Error; err != nil {
		return fmt.Errorf("delete webhook deliveries: %w", err)
	}

	if err := r.db.WithContext(ctx).Delete(&models.WebhookSubscription{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete webhook subscription: %w", err)
	}

	return nil
}

func (r *GormWebhookRepository) CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {

	if delivery == nil {
		return fmt.Errorf("webhook delivery is required")
	}

	return wrap(r.db.WithContext(ctx).Create(delivery).Error, "create webhook delivery")
}

func (r *GormWebhookRepository) UpdateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {

	if delivery == nil {
		return fmt.Errorf("webhook delivery is required")
	}

	return wrap(r.db.WithContext(ctx).Save(delivery).Error, "update webhook delivery")
}

func (r *GormWebhookRepository) ClaimPendingDeliveries(ctx context.Context, now time.Time, limit int) ([]models.WebhookDelivery, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("webhook delivery limit must be between 1 and 100")
	}

	var deliveries []models.WebhookDelivery
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reclaimBefore := now.Add(-10 * time.Minute)
		if err := tx.Preload("Subscription").Where("((status IN ? AND available_at <= ?) OR (status = ? AND updated_at < ?))", []models.WebhookDeliveryStatus{models.WebhookDeliveryPending, models.WebhookDeliveryFailed}, now, models.WebhookDeliverySending, reclaimBefore).
			Order("available_at ASC").Limit(limit).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Find(&deliveries).Error; err != nil {
			return err
		}

		if len(deliveries) == 0 {
			return nil
		}

		ids := make([]string, 0, len(deliveries))
		for i := range deliveries {
			deliveries[i].Status = models.WebhookDeliverySending
			deliveries[i].Attempts++
			ids = append(ids, deliveries[i].ID)
		}

		return tx.Model(&models.WebhookDelivery{}).Where("id IN ?", ids).Updates(map[string]any{
			"status":   models.WebhookDeliverySending,
			"attempts": gorm.Expr("attempts + 1"),
		}).Error
	})

	if err != nil {
		return nil, fmt.Errorf("claim webhook deliveries: %w", err)
	}

	return deliveries, nil
}

func (r *GormWebhookRepository) ListDeliveries(ctx context.Context, subscriptionID string) ([]models.WebhookDelivery, error) {
	var deliveries []models.WebhookDelivery

	if err := r.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Order("created_at DESC").Limit(100).Find(&deliveries).Error; err != nil {
		return nil, fmt.Errorf("list webhook deliveries: %w", err)
	}

	return deliveries, nil
}

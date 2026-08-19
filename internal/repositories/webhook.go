package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type WebhookRepository interface {
	Create(context.Context, *models.WebhookSubscription) error
	FindByID(context.Context, string) (models.WebhookSubscription, error)
	ListByOrganization(context.Context, string) ([]models.WebhookSubscription, error)
	Delete(context.Context, string) error
	CreateDelivery(context.Context, *models.WebhookDelivery) error
	UpdateDelivery(context.Context, *models.WebhookDelivery) error
	ListDeliveries(context.Context, string) ([]models.WebhookDelivery, error)
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

func (r *GormWebhookRepository) FindByID(ctx context.Context, id string) (models.WebhookSubscription, error) {
	var subscription models.WebhookSubscription

	if err := r.db.WithContext(ctx).First(&subscription, "id = ?", id).Error; err != nil {
		return models.WebhookSubscription{}, mapError(err)
	}

	return subscription, nil
}

func (r *GormWebhookRepository) ListByOrganization(ctx context.Context, organizationID string) ([]models.WebhookSubscription, error) {
	var subscriptions []models.WebhookSubscription

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at DESC").Find(&subscriptions).Error; err != nil {
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

func (r *GormWebhookRepository) ListDeliveries(ctx context.Context, subscriptionID string) ([]models.WebhookDelivery, error) {
	var deliveries []models.WebhookDelivery

	if err := r.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Order("created_at DESC").Limit(100).Find(&deliveries).Error; err != nil {
		return nil, fmt.Errorf("list webhook deliveries: %w", err)
	}

	return deliveries, nil
}

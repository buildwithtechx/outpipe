package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type GormAPIKeyRepository struct{ db *gorm.DB }

func NewAPIKeyRepository(db *gorm.DB) (*GormAPIKeyRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormAPIKeyRepository{db: db}, nil
}

func (r *GormAPIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {

	if key == nil {
		return fmt.Errorf("api key is required")
	}

	return wrap(r.db.WithContext(ctx).Create(key).Error, "create api key")
}

func (r *GormAPIKeyRepository) FindByPrefix(ctx context.Context, prefix string) (models.APIKey, error) {
	var key models.APIKey

	if err := r.db.WithContext(ctx).Where("prefix = ?", prefix).First(&key).Error; err != nil {
		return models.APIKey{}, mapError(err)
	}

	return key, nil
}

func (r *GormAPIKeyRepository) FindByID(ctx context.Context, id string) (models.APIKey, error) {
	var key models.APIKey

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error; err != nil {
		return models.APIKey{}, mapError(err)
	}

	return key, nil
}

func (r *GormAPIKeyRepository) ListByUser(ctx context.Context, userID string, organizationID *string) ([]models.APIKey, error) {
	var keys []models.APIKey
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if err := query.Order("created_at DESC").Limit(DefaultListLimit).Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("list user api keys: %w", err)
	}

	return keys, nil
}

func (r *GormAPIKeyRepository) Touch(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.APIKey{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"last_used_at": at}).Error, "touch api key")
}

func (r *GormAPIKeyRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.APIKey{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"revoked_at": at}).Error, "revoke api key")
}

func (r *GormAPIKeyRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at <= ? OR revoked_at IS NOT NULL", now).Delete(&models.APIKey{})
	return result.RowsAffected, wrap(result.Error, "delete expired api keys")
}

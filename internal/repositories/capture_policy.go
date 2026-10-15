package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type CapturePolicyRepository interface {
	Upsert(context.Context, *models.CapturePolicy) error
	FindByTunnel(context.Context, string) (models.CapturePolicy, error)
	DeleteByTunnel(context.Context, string) error
}

type GormCapturePolicyRepository struct{ db *gorm.DB }

func NewCapturePolicyRepository(db *gorm.DB) (*GormCapturePolicyRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormCapturePolicyRepository{db: db}, nil
}

func (r *GormCapturePolicyRepository) Upsert(ctx context.Context, policy *models.CapturePolicy) error {
	if policy == nil || policy.TunnelID == "" || policy.OrganizationID == "" {
		return fmt.Errorf("capture policy tunnel and organization are required")
	}
	var existing models.CapturePolicy
	err := r.db.WithContext(ctx).Where("tunnel_id = ?", policy.TunnelID).First(&existing).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("find capture policy: %w", err)
		}
		policy.Version = 1
		return wrap(r.db.WithContext(ctx).Create(policy).Error, "create capture policy")
	}
	policy.ID = existing.ID
	policy.Version = existing.Version + 1
	return wrap(r.db.WithContext(ctx).Save(policy).Error, "update capture policy")
}

func (r *GormCapturePolicyRepository) FindByTunnel(ctx context.Context, tunnelID string) (models.CapturePolicy, error) {
	var policy models.CapturePolicy
	if err := r.db.WithContext(ctx).Where("tunnel_id = ?", tunnelID).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.CapturePolicy{}, ErrNotFound
		}
		return models.CapturePolicy{}, fmt.Errorf("find capture policy: %w", err)
	}
	return policy, nil
}

func (r *GormCapturePolicyRepository) DeleteByTunnel(ctx context.Context, tunnelID string) error {
	result := r.db.WithContext(ctx).Where("tunnel_id = ?", tunnelID).Delete(&models.CapturePolicy{})
	if result.Error != nil {
		return wrap(result.Error, "delete capture policy")
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type TunnelTokenRepository interface {
	Create(context.Context, *models.TunnelToken) error
	FindByPrefix(context.Context, string) (models.TunnelToken, error)
	Touch(context.Context, string, time.Time) error
	Revoke(context.Context, string, time.Time) error
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type GormTunnelTokenRepository struct{ db *gorm.DB }

func NewTunnelTokenRepository(db *gorm.DB) (*GormTunnelTokenRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormTunnelTokenRepository{db: db}, nil
}

func (r *GormTunnelTokenRepository) Create(ctx context.Context, token *models.TunnelToken) error {

	if token == nil {
		return fmt.Errorf("tunnel token is required")
	}

	return wrap(r.db.WithContext(ctx).Create(token).Error, "create tunnel token")
}

func (r *GormTunnelTokenRepository) FindByPrefix(ctx context.Context, prefix string) (models.TunnelToken, error) {
	var token models.TunnelToken

	if err := r.db.WithContext(ctx).Where("prefix = ?", prefix).First(&token).Error; err != nil {
		return models.TunnelToken{}, mapError(err)
	}

	return token, nil
}

func (r *GormTunnelTokenRepository) Touch(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.TunnelToken{}).Where("id = ? AND revoked_at IS NULL", id).Update("last_used_at", at).Error, "touch tunnel token")
}

func (r *GormTunnelTokenRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.TunnelToken{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", at).Error, "revoke tunnel token")
}

func (r *GormTunnelTokenRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ? OR revoked_at IS NOT NULL", now).Delete(&models.TunnelToken{})
	return result.RowsAffected, wrap(result.Error, "delete expired tunnel tokens")
}

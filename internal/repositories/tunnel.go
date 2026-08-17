package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type TunnelRepository interface {
	Create(context.Context, *models.Tunnel) error
	FindByID(context.Context, string) (models.Tunnel, error)
	FindByHostname(context.Context, string) (models.Tunnel, error)
	FindByOrganization(context.Context, string) ([]models.Tunnel, error)
	CountByOrganization(context.Context, string) (int64, error)
	Update(context.Context, *models.Tunnel) error
	UpdateStatus(context.Context, string, models.TunnelStatus) error
	Touch(context.Context, string, time.Time) error
	Revoke(context.Context, string, time.Time) error
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type GormTunnelRepository struct{ db *gorm.DB }

func NewTunnelRepository(db *gorm.DB) (*GormTunnelRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormTunnelRepository{db: db}, nil
}

func (r *GormTunnelRepository) Create(ctx context.Context, tunnel *models.Tunnel) error {

	if tunnel == nil {
		return fmt.Errorf("tunnel is required")
	}

	return wrap(r.db.WithContext(ctx).Create(tunnel).Error, "create tunnel")
}

func (r *GormTunnelRepository) FindByID(ctx context.Context, id string) (models.Tunnel, error) {
	var tunnel models.Tunnel

	if err := r.db.WithContext(ctx).First(&tunnel, "id = ?", id).Error; err != nil {
		return models.Tunnel{}, mapError(err)
	}

	return tunnel, nil
}

func (r *GormTunnelRepository) FindByHostname(ctx context.Context, hostname string) (models.Tunnel, error) {
	var tunnel models.Tunnel

	if err := r.db.WithContext(ctx).Where("public_hostname = ?", hostname).First(&tunnel).Error; err != nil {
		return models.Tunnel{}, mapError(err)
	}

	return tunnel, nil
}

func (r *GormTunnelRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.Tunnel, error) {
	var tunnels []models.Tunnel

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at DESC").Find(&tunnels).Error; err != nil {
		return nil, fmt.Errorf("list tunnels: %w", err)
	}

	return tunnels, nil
}

func (r *GormTunnelRepository) CountByOrganization(ctx context.Context, organizationID string) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&models.Tunnel{}).Where("organization_id = ? AND revoked_at IS NULL", organizationID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count organization tunnels: %w", err)
	}

	return count, nil
}

func (r *GormTunnelRepository) Update(ctx context.Context, tunnel *models.Tunnel) error {

	if tunnel == nil {
		return fmt.Errorf("tunnel is required")
	}

	return wrap(r.db.WithContext(ctx).Save(tunnel).Error, "update tunnel")
}

func (r *GormTunnelRepository) UpdateStatus(ctx context.Context, id string, status models.TunnelStatus) error {
	result := r.db.WithContext(ctx).Model(&models.Tunnel{}).Where("id = ?", id).Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("update tunnel status: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormTunnelRepository) Touch(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.Tunnel{}).Where("id = ?", id).Updates(map[string]any{"last_active_at": at, "status": models.TunnelStatusActive}).Error, "touch tunnel")
}

func (r *GormTunnelRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.Tunnel{}).Where("id = ?", id).Updates(map[string]any{"status": models.TunnelStatusRevoked, "revoked_at": at}).Error, "revoke tunnel")
}

func (r *GormTunnelRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ?", now).Delete(&models.Tunnel{})
	return result.RowsAffected, wrap(result.Error, "delete expired tunnels")
}

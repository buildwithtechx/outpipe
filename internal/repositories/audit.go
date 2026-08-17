package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type AuditRepository interface {
	Create(context.Context, *models.AuditEvent) error
	ListByOrganization(context.Context, string, time.Time, time.Time, int) ([]models.AuditEvent, error)
	DeleteBefore(context.Context, string, time.Time) (int64, error)
}

type GormAuditRepository struct{ db *gorm.DB }

func NewAuditRepository(db *gorm.DB) (*GormAuditRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormAuditRepository{db: db}, nil
}

func (r *GormAuditRepository) Create(ctx context.Context, event *models.AuditEvent) error {
	if event == nil {
		return fmt.Errorf("audit event is required")
	}
	return wrap(r.db.WithContext(ctx).Create(event).Error, "create audit event")
}

func (r *GormAuditRepository) ListByOrganization(ctx context.Context, organizationID string, from, to time.Time, limit int) ([]models.AuditEvent, error) {
	if limit <= 0 || limit > 1000 {
		return nil, fmt.Errorf("audit limit must be between 1 and 1000")
	}
	var events []models.AuditEvent
	err := r.db.WithContext(ctx).Where("organization_id = ? AND occurred_at >= ? AND occurred_at < ?", organizationID, from, to).Order("occurred_at DESC").Limit(limit).Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	return events, nil
}

func (r *GormAuditRepository) DeleteBefore(ctx context.Context, organizationID string, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("organization_id = ? AND occurred_at < ?", organizationID, before).Delete(&models.AuditEvent{})
	return result.RowsAffected, wrap(result.Error, "delete audit events before retention cutoff")
}

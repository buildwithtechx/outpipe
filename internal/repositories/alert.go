package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type AlertRepository interface {
	Create(context.Context, *models.ObservabilityAlert) error
	FindByID(context.Context, string) (models.ObservabilityAlert, error)
	FindByOrganization(context.Context, string) ([]models.ObservabilityAlert, error)
	Update(context.Context, *models.ObservabilityAlert) error
	Delete(context.Context, string) error
	ClaimDueAlerts(context.Context, time.Time, int) ([]models.ObservabilityAlert, error)
	OpenIncident(context.Context, *models.AlertIncident) error
	GetOpenIncident(context.Context, string) (models.AlertIncident, error)
	ResolveIncident(context.Context, string, time.Time) error
}

type GormAlertRepository struct{ db *gorm.DB }

func NewAlertRepository(db *gorm.DB) (*GormAlertRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormAlertRepository{db: db}, nil
}

func (r *GormAlertRepository) Create(ctx context.Context, alert *models.ObservabilityAlert) error {
	if alert == nil {
		return fmt.Errorf("alert is required")
	}
	return wrap(r.db.WithContext(ctx).Create(alert).Error, "create alert")
}

func (r *GormAlertRepository) FindByID(ctx context.Context, id string) (models.ObservabilityAlert, error) {
	var alert models.ObservabilityAlert
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&alert).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ObservabilityAlert{}, ErrNotFound
		}
		return models.ObservabilityAlert{}, fmt.Errorf("find alert: %w", err)
	}
	return alert, nil
}

func (r *GormAlertRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.ObservabilityAlert, error) {
	var alerts []models.ObservabilityAlert
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return alerts, nil
}

func (r *GormAlertRepository) Update(ctx context.Context, alert *models.ObservabilityAlert) error {
	if alert == nil {
		return fmt.Errorf("alert is required")
	}
	return wrap(r.db.WithContext(ctx).Save(alert).Error, "update alert")
}

func (r *GormAlertRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ObservabilityAlert{})
	if result.Error != nil {
		return wrap(result.Error, "delete alert")
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormAlertRepository) ClaimDueAlerts(ctx context.Context, now time.Time, limit int) ([]models.ObservabilityAlert, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("alert claim limit must be between 1 and 100")
	}
	var alerts []models.ObservabilityAlert
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("enabled = ? AND next_evaluation_at <= ?", true, now).
			Order("next_evaluation_at ASC").Limit(limit).Clauses(clauseLockSkipLocked()).Find(&alerts).Error; err != nil {
			return err
		}
		if len(alerts) == 0 {
			return nil
		}
		ids := make([]string, 0, len(alerts))
		for i := range alerts {
			ids = append(ids, alerts[i].ID)
		}
		return tx.Model(&models.ObservabilityAlert{}).Where("id IN ?", ids).Update("next_evaluation_at", now.Add(time.Minute)).Error
	})
	if err != nil {
		return nil, fmt.Errorf("claim due alerts: %w", err)
	}
	return alerts, nil
}

func (r *GormAlertRepository) OpenIncident(ctx context.Context, link *models.AlertIncident) error {
	if link == nil {
		return fmt.Errorf("alert incident link is required")
	}
	var existing models.AlertIncident
	if err := r.db.WithContext(ctx).Where("alert_id = ? AND resolved_at IS NULL", link.AlertID).First(&existing).Error; err == nil {
		return fmt.Errorf("alert %s already has an open incident %s", link.AlertID, existing.IncidentID)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find open alert incident: %w", err)
	}
	return wrap(r.db.WithContext(ctx).Create(link).Error, "link alert incident")
}

func (r *GormAlertRepository) GetOpenIncident(ctx context.Context, alertID string) (models.AlertIncident, error) {
	var link models.AlertIncident
	if err := r.db.WithContext(ctx).Where("alert_id = ? AND resolved_at IS NULL", alertID).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.AlertIncident{}, ErrNotFound
		}
		return models.AlertIncident{}, fmt.Errorf("find open alert incident: %w", err)
	}
	return link, nil
}

func (r *GormAlertRepository) ResolveIncident(ctx context.Context, alertID string, resolvedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.AlertIncident{}).Where("alert_id = ? AND resolved_at IS NULL", alertID).Update("resolved_at", resolvedAt)
	if result.Error != nil {
		return wrap(result.Error, "resolve alert incident")
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

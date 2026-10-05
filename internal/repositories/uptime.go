package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type UptimeRepository interface {
	CreateMonitor(ctx context.Context, monitor *models.UptimeMonitor) error
	GetMonitor(ctx context.Context, id string) (*models.UptimeMonitor, error)
	ListOrgMonitors(ctx context.Context, orgID string) ([]models.UptimeMonitor, error)
	UpdateMonitor(ctx context.Context, monitor *models.UptimeMonitor) error
	DeleteMonitor(ctx context.Context, id string) error
	RecordCheck(ctx context.Context, check *models.UptimeCheck) error
	GetRecentChecks(ctx context.Context, monitorID string, limit int) ([]models.UptimeCheck, error)

	CreateIncident(ctx context.Context, incident *models.UptimeIncident) error
	GetIncident(ctx context.Context, id string) (*models.UptimeIncident, error)
	ListOrgIncidents(ctx context.Context, orgID string) ([]models.UptimeIncident, error)
	UpdateIncident(ctx context.Context, incident *models.UptimeIncident) error
	AddIncidentUpdate(ctx context.Context, update *models.UptimeIncidentUpdate) error

	GetStatusPageByOrg(ctx context.Context, orgID string) (*models.UptimeStatusPage, error)
	GetStatusPageBySlug(ctx context.Context, slug string) (*models.UptimeStatusPage, error)
	UpsertStatusPage(ctx context.Context, page *models.UptimeStatusPage) error
	AddSubscriber(ctx context.Context, subscriber *models.UptimeSubscriber) error
	ListSubscribers(ctx context.Context, statusPageID string) ([]models.UptimeSubscriber, error)
}

type GormUptimeRepository struct {
	db *gorm.DB
}

func NewUptimeRepository(db *gorm.DB) (*GormUptimeRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	return &GormUptimeRepository{db: db}, nil
}

func (r *GormUptimeRepository) CreateMonitor(ctx context.Context, monitor *models.UptimeMonitor) error {
	if err := r.db.WithContext(ctx).Create(monitor).Error; err != nil {
		return fmt.Errorf("create uptime monitor: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) GetMonitor(ctx context.Context, id string) (*models.UptimeMonitor, error) {
	var monitor models.UptimeMonitor
	if err := r.db.WithContext(ctx).First(&monitor, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("monitor %q not found: %w", id, err)
		}
		return nil, fmt.Errorf("find monitor %q: %w", id, err)
	}
	return &monitor, nil
}

func (r *GormUptimeRepository) ListOrgMonitors(ctx context.Context, orgID string) ([]models.UptimeMonitor, error) {
	var monitors []models.UptimeMonitor
	if err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("created_at desc").Find(&monitors).Error; err != nil {
		return nil, fmt.Errorf("list org monitors: %w", err)
	}
	return monitors, nil
}

func (r *GormUptimeRepository) UpdateMonitor(ctx context.Context, monitor *models.UptimeMonitor) error {
	result := r.db.WithContext(ctx).Model(&models.UptimeMonitor{}).Where("id = ?", monitor.ID).Select("*").Updates(monitor)
	if result.Error != nil {
		return fmt.Errorf("update monitor %q: %w", monitor.ID, result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *GormUptimeRepository) DeleteMonitor(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.UptimeMonitor{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete monitor %q: %w", id, err)
	}
	return nil
}

func (r *GormUptimeRepository) RecordCheck(ctx context.Context, check *models.UptimeCheck) error {
	if err := r.db.WithContext(ctx).Create(check).Error; err != nil {
		return fmt.Errorf("record uptime check: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) GetRecentChecks(ctx context.Context, monitorID string, limit int) ([]models.UptimeCheck, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var checks []models.UptimeCheck
	if err := r.db.WithContext(ctx).Where("monitor_id = ?", monitorID).Order("checked_at desc").Limit(limit).Find(&checks).Error; err != nil {
		return nil, fmt.Errorf("get recent checks for monitor %q: %w", monitorID, err)
	}
	return checks, nil
}

func (r *GormUptimeRepository) CreateIncident(ctx context.Context, incident *models.UptimeIncident) error {
	if err := r.db.WithContext(ctx).Create(incident).Error; err != nil {
		return fmt.Errorf("create incident: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) GetIncident(ctx context.Context, id string) (*models.UptimeIncident, error) {
	var incident models.UptimeIncident
	if err := r.db.WithContext(ctx).Preload("Updates", func(db *gorm.DB) *gorm.DB {
		return db.Order("uptime_incident_updates.created_at asc")
	}).First(&incident, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("incident %q not found: %w", id, err)
		}
		return nil, fmt.Errorf("find incident %q: %w", id, err)
	}
	return &incident, nil
}

func (r *GormUptimeRepository) ListOrgIncidents(ctx context.Context, orgID string) ([]models.UptimeIncident, error) {
	var incidents []models.UptimeIncident
	if err := r.db.WithContext(ctx).Preload("Updates", func(db *gorm.DB) *gorm.DB {
		return db.Order("uptime_incident_updates.created_at asc")
	}).Where("organization_id = ?", orgID).Order("created_at desc").Find(&incidents).Error; err != nil {
		return nil, fmt.Errorf("list org incidents: %w", err)
	}
	return incidents, nil
}

func (r *GormUptimeRepository) UpdateIncident(ctx context.Context, incident *models.UptimeIncident) error {
	if err := r.db.WithContext(ctx).Save(incident).Error; err != nil {
		return fmt.Errorf("update incident %q: %w", incident.ID, err)
	}
	return nil
}

func (r *GormUptimeRepository) AddIncidentUpdate(ctx context.Context, update *models.UptimeIncidentUpdate) error {
	if err := r.db.WithContext(ctx).Create(update).Error; err != nil {
		return fmt.Errorf("add incident update: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) GetStatusPageByOrg(ctx context.Context, orgID string) (*models.UptimeStatusPage, error) {
	var page models.UptimeStatusPage
	if err := r.db.WithContext(ctx).First(&page, "organization_id = ?", orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("status page for org %q not found: %w", orgID, err)
		}
		return nil, fmt.Errorf("find status page for org %q: %w", orgID, err)
	}
	return &page, nil
}

func (r *GormUptimeRepository) GetStatusPageBySlug(ctx context.Context, slug string) (*models.UptimeStatusPage, error) {
	var page models.UptimeStatusPage
	result := r.db.WithContext(ctx).Where("slug = ?", slug).Limit(1).Find(&page)
	if result.Error != nil {
		return nil, fmt.Errorf("find status page %q: %w", slug, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("status page %q not found: %w", slug, gorm.ErrRecordNotFound)
	}
	return &page, nil
}

func (r *GormUptimeRepository) UpsertStatusPage(ctx context.Context, page *models.UptimeStatusPage) error {
	var existing models.UptimeStatusPage
	err := r.db.WithContext(ctx).First(&existing, "organization_id = ?", page.OrganizationID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check existing status page: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.WithContext(ctx).Create(page).Error; err != nil {
			return fmt.Errorf("create status page: %w", err)
		}
		return nil
	}

	existing.Slug = page.Slug
	existing.Title = page.Title
	existing.Description = page.Description
	existing.CustomDomain = page.CustomDomain
	existing.Published = page.Published
	existing.UpdatedAt = page.UpdatedAt

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return fmt.Errorf("update status page: %w", err)
	}
	*page = existing
	return nil
}

func (r *GormUptimeRepository) AddSubscriber(ctx context.Context, subscriber *models.UptimeSubscriber) error {
	if err := r.db.WithContext(ctx).Create(subscriber).Error; err != nil {
		return fmt.Errorf("add subscriber: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) ListSubscribers(ctx context.Context, statusPageID string) ([]models.UptimeSubscriber, error) {
	var subscribers []models.UptimeSubscriber
	if err := r.db.WithContext(ctx).Where("status_page_id = ?", statusPageID).Find(&subscribers).Error; err != nil {
		return nil, fmt.Errorf("list subscribers: %w", err)
	}
	return subscribers, nil
}

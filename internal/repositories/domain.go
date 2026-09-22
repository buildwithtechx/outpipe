package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type DomainRepository interface {
	Create(context.Context, *models.Domain) error
	FindByID(context.Context, string) (models.Domain, error)
	FindByHostname(context.Context, string) (models.Domain, error)
	FindByOrganization(context.Context, string) ([]models.Domain, error)
	CountByOrganization(context.Context, string) (int64, error)
	Update(context.Context, *models.Domain) error
	UpdateStatus(context.Context, string, models.DomainStatus) error
	Delete(context.Context, string) error
}

type GormDomainRepository struct{ db *gorm.DB }

func NewDomainRepository(db *gorm.DB) (*GormDomainRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormDomainRepository{db: db}, nil
}

func (r *GormDomainRepository) Create(ctx context.Context, domain *models.Domain) error {

	if domain == nil {
		return fmt.Errorf("domain is required")
	}

	return wrap(r.db.WithContext(ctx).Create(domain).Error, "create domain")
}

func (r *GormDomainRepository) FindByID(ctx context.Context, id string) (models.Domain, error) {
	var domain models.Domain

	if err := r.db.WithContext(ctx).First(&domain, "id = ?", id).Error; err != nil {
		return models.Domain{}, mapError(err)
	}

	return domain, nil
}

func (r *GormDomainRepository) FindByHostname(ctx context.Context, hostname string) (models.Domain, error) {
	var domain models.Domain

	if err := r.db.WithContext(ctx).Where("hostname = ?", hostname).First(&domain).Error; err != nil {
		return models.Domain{}, mapError(err)
	}

	return domain, nil
}

func (r *GormDomainRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.Domain, error) {
	var domains []models.Domain

	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at DESC").Limit(DefaultListLimit).Find(&domains).Error; err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}

	return domains, nil
}

func (r *GormDomainRepository) CountByOrganization(ctx context.Context, organizationID string) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&models.Domain{}).Where("organization_id = ? AND status != ?", organizationID, models.DomainStatusRevoked).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count organization domains: %w", err)
	}

	return count, nil
}

func (r *GormDomainRepository) Update(ctx context.Context, domain *models.Domain) error {

	if domain == nil {
		return fmt.Errorf("domain is required")
	}

	return wrap(r.db.WithContext(ctx).Save(domain).Error, "update domain")
}

func (r *GormDomainRepository) UpdateStatus(ctx context.Context, id string, status models.DomainStatus) error {
	result := r.db.WithContext(ctx).Model(&models.Domain{}).Where("id = ?", id).Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("update domain status: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormDomainRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Domain{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("delete domain: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

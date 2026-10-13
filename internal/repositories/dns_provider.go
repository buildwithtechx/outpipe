package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type DNSProviderRepository interface {
	Upsert(context.Context, *models.DNSProviderCredential) error
	FindByID(context.Context, string) (models.DNSProviderCredential, error)
	FindByOrganization(context.Context, string) ([]models.DNSProviderCredential, error)
	Delete(context.Context, string) error
}

type GormDNSProviderRepository struct{ db *gorm.DB }

func NewDNSProviderRepository(db *gorm.DB) (*GormDNSProviderRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormDNSProviderRepository{db: db}, nil
}

func (r *GormDNSProviderRepository) Upsert(ctx context.Context, credential *models.DNSProviderCredential) error {
	if credential == nil || credential.OrganizationID == "" || credential.Provider == "" || credential.Ciphertext == "" {
		return fmt.Errorf("complete DNS provider credential is required")
	}
	var existing models.DNSProviderCredential
	err := r.db.WithContext(ctx).Where("organization_id = ? AND provider = ?", credential.OrganizationID, credential.Provider).First(&existing).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("find DNS provider credential: %w", err)
		}
		return wrap(r.db.WithContext(ctx).Create(credential).Error, "create DNS provider credential")
	}
	credential.ID = existing.ID
	return wrap(r.db.WithContext(ctx).Save(credential).Error, "rotate DNS provider credential")
}

func (r *GormDNSProviderRepository) FindByID(ctx context.Context, id string) (models.DNSProviderCredential, error) {
	var credential models.DNSProviderCredential
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&credential).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.DNSProviderCredential{}, ErrNotFound
		}
		return models.DNSProviderCredential{}, fmt.Errorf("find DNS provider credential: %w", err)
	}
	return credential, nil
}

func (r *GormDNSProviderRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.DNSProviderCredential, error) {
	var credentials []models.DNSProviderCredential
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("provider ASC").Find(&credentials).Error; err != nil {
		return nil, fmt.Errorf("list DNS provider credentials: %w", err)
	}
	return credentials, nil
}

func (r *GormDNSProviderRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.DNSProviderCredential{})
	if result.Error != nil {
		return wrap(result.Error, "delete DNS provider credential")
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

type OrganizationRepository interface {
	Create(context.Context, *models.Organization) error
	FindByID(context.Context, string) (models.Organization, error)
	FindBySlug(context.Context, string) (models.Organization, error)
	IsSlugAvailable(context.Context, string) (bool, error)
	List(context.Context) ([]models.Organization, error)
	Update(context.Context, *models.Organization) error
	AddMember(context.Context, *models.OrganizationMember) error
	AddMemberWithLimit(context.Context, *models.OrganizationMember, int64) error
	FindMember(context.Context, string, string) (models.OrganizationMember, error)
	ListMembers(context.Context, string) ([]models.OrganizationMember, error)
	RemoveMember(context.Context, string, string) error
	ListOwned(context.Context, string) ([]models.Organization, error)
	ListForUser(context.Context, string) ([]models.Organization, error)
	TransferOwnership(context.Context, string, string, string) error
	CountMembers(context.Context, string) (int64, error)
}

type GormOrganizationRepository struct{ db *gorm.DB }

func NewOrganizationRepository(db *gorm.DB) (*GormOrganizationRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormOrganizationRepository{db: db}, nil
}

func (r *GormOrganizationRepository) Create(ctx context.Context, organization *models.Organization) error {

	if organization == nil {
		return fmt.Errorf("organization is required")
	}

	return wrap(r.db.WithContext(ctx).Create(organization).Error, "create organization")
}

func (r *GormOrganizationRepository) FindByID(ctx context.Context, id string) (models.Organization, error) {
	var organization models.Organization
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&organization).Error

	if err != nil {
		return models.Organization{}, mapError(err)
	}

	return organization, nil
}

func (r *GormOrganizationRepository) FindBySlug(ctx context.Context, slug string) (models.Organization, error) {
	var organization models.Organization
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&organization).Error

	if err != nil {
		return models.Organization{}, mapError(err)
	}

	return organization, nil
}

func (r *GormOrganizationRepository) IsSlugAvailable(ctx context.Context, slug string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&models.Organization{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check organization slug availability: %w", err)
	}

	return count == 0, nil
}

func (r *GormOrganizationRepository) List(ctx context.Context) ([]models.Organization, error) {
	var organizations []models.Organization

	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&organizations).Error; err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}

	return organizations, nil
}

func (r *GormOrganizationRepository) Update(ctx context.Context, organization *models.Organization) error {

	if organization == nil {
		return fmt.Errorf("organization is required")
	}

	return wrap(r.db.WithContext(ctx).Save(organization).Error, "update organization")
}

func (r *GormOrganizationRepository) AddMember(ctx context.Context, member *models.OrganizationMember) error {

	if member == nil {
		return fmt.Errorf("organization member is required")
	}

	return wrap(r.db.WithContext(ctx).Create(member).Error, "add organization member")
}

func (r *GormOrganizationRepository) AddMemberWithLimit(ctx context.Context, member *models.OrganizationMember, limit int64) error {

	if member == nil {
		return fmt.Errorf("organization member is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var organization models.Organization

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", member.OrganizationID).First(&organization).Error; err != nil {
			return mapError(err)
		}

		if limit > 0 {
			var count int64

			if err := tx.Model(&models.OrganizationMember{}).Where("organization_id = ?", member.OrganizationID).Count(&count).Error; err != nil {
				return fmt.Errorf("count organization members: %w", err)
			}

			if count >= limit {
				return fmt.Errorf("organization member limit reached")
			}
		}

		return wrap(tx.Create(member).Error, "add organization member")
	})
}

func (r *GormOrganizationRepository) FindMember(ctx context.Context, organizationID, userID string) (models.OrganizationMember, error) {
	var member models.OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", organizationID, userID).First(&member).Error

	if err != nil {
		return models.OrganizationMember{}, mapError(err)
	}

	return member, nil
}

func (r *GormOrganizationRepository) ListMembers(ctx context.Context, organizationID string) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Order("created_at ASC").Find(&members).Error

	if err != nil {
		return nil, fmt.Errorf("list organization members: %w", err)
	}

	return members, nil
}

func (r *GormOrganizationRepository) RemoveMember(ctx context.Context, organizationID, userID string) error {
	result := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", organizationID, userID).Delete(&models.OrganizationMember{})

	if result.Error != nil {
		return fmt.Errorf("remove organization member: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormOrganizationRepository) ListOwned(ctx context.Context, ownerID string) ([]models.Organization, error) {
	var organizations []models.Organization

	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&organizations).Error; err != nil {
		return nil, fmt.Errorf("list owned organizations: %w", err)
	}

	return organizations, nil
}

func (r *GormOrganizationRepository) ListForUser(ctx context.Context, userID string) ([]models.Organization, error) {
	var organizations []models.Organization
	err := r.db.WithContext(ctx).
		Table("organizations").
		Select("organizations.*").
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userID).
		Order("organizations.name ASC").
		Find(&organizations).Error

	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}

	return organizations, nil
}

func (r *GormOrganizationRepository) TransferOwnership(ctx context.Context, organizationID, currentOwnerID, newOwnerID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var member models.OrganizationMember

		if err := tx.Where("organization_id = ? AND user_id = ?", organizationID, newOwnerID).First(&member).Error; err != nil {
			return mapError(err)
		}

		result := tx.Model(&models.Organization{}).Where("id = ? AND owner_id = ?", organizationID, currentOwnerID).Update("owner_id", newOwnerID)

		if result.Error != nil {
			return fmt.Errorf("transfer organization ownership: %w", result.Error)
		}

		if result.RowsAffected != 1 {
			return ErrNotFound
		}

		if err := tx.Model(&models.OrganizationMember{}).Where("organization_id = ? AND user_id = ?", organizationID, currentOwnerID).Update("role", models.MemberRoleAdmin).Error; err != nil {
			return fmt.Errorf("update previous owner role: %w", err)
		}

		if err := tx.Model(&models.OrganizationMember{}).Where("organization_id = ? AND user_id = ?", organizationID, newOwnerID).Update("role", models.MemberRoleOwner).Error; err != nil {
			return fmt.Errorf("update new owner role: %w", err)
		}

		return nil
	})
}

func (r *GormOrganizationRepository) CountMembers(ctx context.Context, organizationID string) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&models.OrganizationMember{}).Where("organization_id = ?", organizationID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count organization members: %w", err)
	}

	return count, nil
}

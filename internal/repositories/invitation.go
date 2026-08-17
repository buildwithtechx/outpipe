package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

type OrganizationInvitationRepository interface {
	CreateInvitation(context.Context, *models.OrganizationInvitation) error
	DeleteInvitation(context.Context, string) error
	FindActiveInvitation(context.Context, string, time.Time) (models.OrganizationInvitation, error)
	AcceptInvitation(context.Context, string, string, string, models.MemberRole, int64, time.Time) error
}

func (r *GormOrganizationInvitationRepository) DeleteInvitation(ctx context.Context, id string) error {
	return wrap(r.db.WithContext(ctx).Where("id = ? AND accepted_at IS NULL", id).Delete(&models.OrganizationInvitation{}).Error, "delete organization invitation")
}

type GormOrganizationInvitationRepository struct{ db *gorm.DB }

func NewOrganizationInvitationRepository(db *gorm.DB) (*GormOrganizationInvitationRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormOrganizationInvitationRepository{db: db}, nil
}

func (r *GormOrganizationInvitationRepository) CreateInvitation(ctx context.Context, invitation *models.OrganizationInvitation) error {

	if invitation == nil {
		return fmt.Errorf("organization invitation is required")
	}

	return wrap(r.db.WithContext(ctx).Create(invitation).Error, "create organization invitation")
}

func (r *GormOrganizationInvitationRepository) FindActiveInvitation(ctx context.Context, tokenHash string, now time.Time) (models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	err := r.db.WithContext(ctx).Where("token_hash = ? AND expires_at > ? AND accepted_at IS NULL AND revoked_at IS NULL", tokenHash, now).First(&invitation).Error

	if err != nil {
		return models.OrganizationInvitation{}, mapError(err)
	}

	return invitation, nil
}

func (r *GormOrganizationInvitationRepository) AcceptInvitation(ctx context.Context, invitationID, organizationID, userID string, role models.MemberRole, memberLimit int64, at time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var organization models.Organization

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", organizationID).First(&organization).Error; err != nil {
			return mapError(err)
		}

		var invitation models.OrganizationInvitation

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ? AND expires_at > ? AND accepted_at IS NULL AND revoked_at IS NULL", invitationID, organizationID, at).First(&invitation).Error; err != nil {
			return mapError(err)
		}

		if memberLimit > 0 {
			var members int64

			if err := tx.Model(&models.OrganizationMember{}).Where("organization_id = ?", organizationID).Count(&members).Error; err != nil {
				return fmt.Errorf("count organization members: %w", err)
			}

			if members >= memberLimit {
				return fmt.Errorf("organization member limit reached")
			}
		}

		member := &models.OrganizationMember{OrganizationID: organizationID, UserID: userID, Role: role}

		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("create invited member: %w", err)
		}

		if err := tx.Model(&models.OrganizationInvitation{}).Where("id = ?", invitationID).Update("accepted_at", at).Error; err != nil {
			return fmt.Errorf("mark invitation accepted: %w", err)
		}

		return nil
	})
}

package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/pkg/utils"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

type OrganizationService struct {
	organizations repositories.OrganizationRepository
	billing       *BillingService
}

func (s *OrganizationService) SetBilling(billing *BillingService) { s.billing = billing }

func NewOrganizationService(organizations repositories.OrganizationRepository) (*OrganizationService, error) {

	if organizations == nil {
		return nil, fmt.Errorf("organization repository is required")
	}

	return &OrganizationService{organizations: organizations}, nil
}

func (s *OrganizationService) Create(ctx context.Context, ownerID, name, slug string) (models.Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))

	if ownerID == "" || name == "" || !slugPattern.MatchString(slug) {
		return models.Organization{}, fmt.Errorf("owner, name, and valid slug are required")
	}

	organization := models.Organization{Name: name, Slug: slug, OwnerID: ownerID, Settings: `{}`}

	if err := s.organizations.Create(ctx, &organization); err != nil {
		return models.Organization{}, fmt.Errorf("create organization: %w", err)
	}

	member := models.OrganizationMember{OrganizationID: organization.ID, UserID: ownerID, Role: models.MemberRoleOwner}

	if err := s.organizations.AddMember(ctx, &member); err != nil {
		return models.Organization{}, fmt.Errorf("add organization owner: %w", err)
	}

	if s.billing != nil {
		if err := s.billing.ProvisionFreeSubscription(ctx, organization.ID); err != nil {
			return models.Organization{}, fmt.Errorf("provision free subscription: %w", err)
		}
	}

	return organization, nil
}

func (s *OrganizationService) ListForUser(ctx context.Context, userID string) ([]models.Organization, error) {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	organizations, err := s.organizations.ListForUser(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("list user organizations: %w", err)
	}

	return organizations, nil
}

func (s *OrganizationService) Get(ctx context.Context, organizationID string) (models.Organization, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return models.Organization{}, fmt.Errorf("organization id is required")
	}

	organization, err := s.organizations.FindByID(ctx, organizationID)

	if err != nil {
		return models.Organization{}, err
	}

	return organization, nil
}

func (s *OrganizationService) ListMembers(ctx context.Context, organizationID string) ([]models.OrganizationMember, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return nil, fmt.Errorf("organization id is required")
	}

	members, err := s.organizations.ListMembers(ctx, organizationID)

	if err != nil {
		return nil, fmt.Errorf("list organization members: %w", err)
	}

	return members, nil
}

func (s *OrganizationService) RemoveMember(ctx context.Context, organizationID, userID string) error {
	organizationID = strings.TrimSpace(organizationID)
	userID = strings.TrimSpace(userID)

	if organizationID == "" || userID == "" {
		return fmt.Errorf("organization and member are required")
	}

	member, err := s.organizations.FindMember(ctx, organizationID, userID)

	if err != nil {
		return err
	}

	if member.Role == models.MemberRoleOwner {
		return utils.NewAuthorizationError(fmt.Errorf("organization owner cannot be removed"))
	}

	if err := s.organizations.RemoveMember(ctx, organizationID, userID); err != nil {
		return fmt.Errorf("remove organization member: %w", err)
	}

	return nil
}

func (s *OrganizationService) IsSlugAvailable(ctx context.Context, slug string) (bool, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))

	if !slugPattern.MatchString(slug) {
		return false, fmt.Errorf("slug must use lowercase letters, numbers, and single hyphens")
	}

	available, err := s.organizations.IsSlugAvailable(ctx, slug)

	if err != nil {
		return false, fmt.Errorf("check organization slug: %w", err)
	}

	return available, nil
}

func (s *OrganizationService) AddMember(ctx context.Context, organizationID, userID string, role models.MemberRole) error {

	if organizationID == "" || userID == "" || !validMemberRole(role) {
		return fmt.Errorf("organization, user, and valid role are required")
	}

	member := &models.OrganizationMember{OrganizationID: organizationID, UserID: userID, Role: role}
	limit, err := s.memberLimit(ctx, organizationID)

	if err != nil {
		return err
	}

	if err := s.organizations.AddMemberWithLimit(ctx, member, limit); err != nil {
		return fmt.Errorf("add organization member: %w", err)
	}

	return nil
}

func (s *OrganizationService) memberLimit(ctx context.Context, organizationID string) (int64, error) {

	if s.billing == nil {
		return 0, nil
	}

	plan, _, err := s.billing.Entitlements(ctx, organizationID)

	if err != nil {
		return 0, fmt.Errorf("check member entitlement: %w", err)
	}

	return int64(plan.MaxMembers), nil
}

func (s *OrganizationService) Authorize(ctx context.Context, organizationID, userID string, required models.MemberRole) error {
	member, err := s.organizations.FindMember(ctx, organizationID, userID)

	if err != nil {

		if errors.Is(err, repositories.ErrNotFound) {
			return utils.NewAuthorizationError(fmt.Errorf("organization membership required"))
		}

		return fmt.Errorf("find organization membership: %w", err)
	}

	if memberRoleRank(member.Role) < memberRoleRank(required) {
		return utils.NewAuthorizationError(fmt.Errorf("insufficient organization role"))
	}

	return nil
}

func validMemberRole(role models.MemberRole) bool {
	return role == models.MemberRoleOwner || role == models.MemberRoleAdmin || role == models.MemberRoleMember || role == models.MemberRoleViewer
}

func memberRoleRank(role models.MemberRole) int {

	switch role {
	case models.MemberRoleOwner:
		return 4
	case models.MemberRoleAdmin:
		return 3
	case models.MemberRoleMember:
		return 2
	case models.MemberRoleViewer:
		return 1
	default:
		return 0
	}
}

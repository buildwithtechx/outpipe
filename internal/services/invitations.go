package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/pkg/utils"
)

type OrganizationInvitationMailer interface {
	SendOrganizationInvite(context.Context, string, string, string, string, string) error
}

type InvitationService struct {
	invitations   repositories.OrganizationInvitationRepository
	organizations *OrganizationService
	users         repositories.UserRepository
	dashboardURL  string
	tokenTTL      time.Duration
	now           func() time.Time
	mailer        OrganizationInvitationMailer
}

func NewInvitationService(invitations repositories.OrganizationInvitationRepository, organizations *OrganizationService, users repositories.UserRepository, dashboardURL string, tokenTTL time.Duration) (*InvitationService, error) {

	if invitations == nil || organizations == nil || users == nil || strings.TrimSpace(dashboardURL) == "" {
		return nil, fmt.Errorf("invitation dependencies are required")
	}

	if tokenTTL <= 0 {
		return nil, fmt.Errorf("invitation token ttl must be positive")
	}

	return &InvitationService{invitations: invitations, organizations: organizations, users: users, dashboardURL: strings.TrimRight(dashboardURL, "/"), tokenTTL: tokenTTL, now: time.Now}, nil
}

func (s *InvitationService) SetMailer(mailer OrganizationInvitationMailer) { s.mailer = mailer }

func (s *InvitationService) Invite(ctx context.Context, inviterID, organizationID, email string, role models.MemberRole) (models.OrganizationInvitation, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if inviterID == "" || organizationID == "" || email == "" || role == models.MemberRoleOwner || !validMemberRole(role) {
		return models.OrganizationInvitation{}, utils.NewClientErrorf("inviter, organization, email, and valid non-owner role are required")
	}

	if s.mailer == nil {
		return models.OrganizationInvitation{}, fmt.Errorf("invitation delivery is unavailable")
	}

	if err := s.organizations.Authorize(ctx, organizationID, inviterID, models.MemberRoleAdmin); err != nil {
		return models.OrganizationInvitation{}, err
	}

	organization, err := s.organizations.organizations.FindByID(ctx, organizationID)

	if err != nil {
		return models.OrganizationInvitation{}, fmt.Errorf("find organization: %w", err)
	}

	inviter, err := s.users.FindByID(ctx, inviterID)

	if err != nil {
		return models.OrganizationInvitation{}, fmt.Errorf("find inviter: %w", err)
	}

	if existing, findErr := s.users.FindByEmail(ctx, email); findErr == nil {

		if _, memberErr := s.organizations.organizations.FindMember(ctx, organizationID, existing.ID); memberErr == nil {
			return models.OrganizationInvitation{}, utils.NewClientErrorf("user is already an organization member")
		}

	} else if findErr != repositories.ErrNotFound {
		return models.OrganizationInvitation{}, fmt.Errorf("find invited user: %w", findErr)
	}

	rawToken, err := auth.NewToken("cdi", 32)

	if err != nil {
		return models.OrganizationInvitation{}, fmt.Errorf("create invitation token: %w", err)
	}

	now := s.now()
	invitation := models.OrganizationInvitation{OrganizationID: organizationID, InviterID: inviterID, Email: email, Role: role, TokenHash: auth.HashToken(rawToken), ExpiresAt: now.Add(s.tokenTTL)}

	if err := s.invitations.CreateInvitation(ctx, &invitation); err != nil {
		return models.OrganizationInvitation{}, err
	}

	link := s.dashboardURL + "/invitations/accept?token=" + url.QueryEscape(rawToken)

	if err := s.mailer.SendOrganizationInvite(ctx, email, inviter.Name, organization.Name, string(role), link); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if deleteErr := s.invitations.DeleteInvitation(cleanupCtx, invitation.ID); deleteErr != nil {
			return models.OrganizationInvitation{}, fmt.Errorf("send organization invitation: %w; clean up invitation: %v", err, deleteErr)
		}

		return models.OrganizationInvitation{}, fmt.Errorf("send organization invitation: %w", err)
	}

	return invitation, nil
}

func (s *InvitationService) Accept(ctx context.Context, userID, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)

	if userID == "" || rawToken == "" {
		return fmt.Errorf("user and invitation token are required")
	}

	user, err := s.users.FindByID(ctx, userID)

	if err != nil {
		return fmt.Errorf("find accepting user: %w", err)
	}

	invitation, err := s.invitations.FindActiveInvitation(ctx, auth.HashToken(rawToken), s.now())

	if err != nil {
		return fmt.Errorf("find active invitation: %w", err)
	}

	if !strings.EqualFold(user.Email, invitation.Email) {
		return fmt.Errorf("invitation email does not match authenticated user")
	}

	memberLimit, err := s.organizations.memberLimit(ctx, invitation.OrganizationID)

	if err != nil {
		return err
	}

	if err := s.invitations.AcceptInvitation(ctx, invitation.ID, invitation.OrganizationID, userID, invitation.Role, memberLimit, s.now()); err != nil {
		return fmt.Errorf("accept organization invitation: %w", err)
	}

	return nil
}

package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type AccountService struct {
	users         repositories.UserRepository
	organizations repositories.OrganizationRepository
	now           func() time.Time
	mailer        AccountMailer
}

type AccountMailer interface {
	SendAccountUpdate(context.Context, string, string) error
	SendWelcome(context.Context, string, string) error
}

func NewAccountService(users repositories.UserRepository, organizations repositories.OrganizationRepository) (*AccountService, error) {

	if users == nil || organizations == nil {
		return nil, fmt.Errorf("account repositories are required")
	}

	return &AccountService{users: users, organizations: organizations, now: time.Now}, nil
}

func (s *AccountService) SetMailer(mailer AccountMailer) { s.mailer = mailer }

func (s *AccountService) Profile(ctx context.Context, userID string) (models.User, error) {

	if userID == "" {
		return models.User{}, fmt.Errorf("user id is required")
	}

	user, err := s.users.FindByID(ctx, userID)

	if err != nil {
		return models.User{}, fmt.Errorf("find account profile: %w", err)
	}

	return user, nil
}

func (s *AccountService) Delete(ctx context.Context, userID string) error {
	user, err := s.users.FindByID(ctx, userID)

	if err != nil {
		return fmt.Errorf("find account: %w", err)
	}

	owned, err := s.organizations.ListOwned(ctx, userID)

	if err != nil {
		return fmt.Errorf("check owned organizations: %w", err)
	}

	if len(owned) > 0 {
		return fmt.Errorf("transfer organization ownership before deleting account")
	}

	if err := s.users.Delete(ctx, userID, s.now()); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}

	if s.mailer != nil {

		if err := s.mailer.SendAccountUpdate(ctx, user.Email, "deleted"); err != nil {
			return fmt.Errorf("send account update: %w", err)
		}
	}

	return nil
}

func (s *AccountService) TransferOwnership(ctx context.Context, organizationID, currentOwnerID, newOwnerID string) error {

	if organizationID == "" || currentOwnerID == "" || newOwnerID == "" || currentOwnerID == newOwnerID {
		return fmt.Errorf("valid ownership transfer users are required")
	}

	if err := s.organizations.TransferOwnership(ctx, organizationID, currentOwnerID, newOwnerID); err != nil {
		return fmt.Errorf("transfer ownership: %w", err)
	}

	return nil
}

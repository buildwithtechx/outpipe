package services

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/pkg/utils"
)

type AdminService struct{ admin repositories.AdminRepository }

type AdminOverview struct {
	Users         int64 `json:"users"`
	Organizations int64 `json:"organizations"`
	Tunnels       int64 `json:"tunnels"`
	Subscriptions int64 `json:"subscriptions"`
}

func NewAdminService(admin repositories.AdminRepository) (*AdminService, error) {

	if admin == nil {
		return nil, fmt.Errorf("admin repository is required")
	}

	return &AdminService{admin: admin}, nil
}

func (s *AdminService) Overview(ctx context.Context) (AdminOverview, error) {
	users, err := s.admin.CountUsers(ctx)

	if err != nil {
		return AdminOverview{}, err
	}

	organizations, err := s.admin.CountOrganizations(ctx)

	if err != nil {
		return AdminOverview{}, err
	}

	tunnels, err := s.admin.CountTunnels(ctx)

	if err != nil {
		return AdminOverview{}, err
	}

	subscriptions, err := s.admin.CountSubscriptions(ctx)

	if err != nil {
		return AdminOverview{}, err
	}

	return AdminOverview{Users: users, Organizations: organizations, Tunnels: tunnels, Subscriptions: subscriptions}, nil
}

func (s *AdminService) User(ctx context.Context, id string) (models.User, error) {
	if id == "" { return models.User{}, fmt.Errorf("user id is required") }
	return s.admin.FindUser(ctx, id)
}

func (s *AdminService) Organization(ctx context.Context, id string) (models.Organization, error) {
	if id == "" { return models.Organization{}, fmt.Errorf("organization id is required") }
	return s.admin.FindOrganization(ctx, id)
}

func (s *AdminService) Users(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	limit, offset, err := normalizePage(limit, offset)

	if err != nil {
		return nil, 0, utils.ClientError{Err: err}
	}

	users, err := s.admin.ListUsers(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.admin.CountUsers(ctx)
	return users, total, err
}

func (s *AdminService) Organizations(ctx context.Context, limit, offset int) ([]models.Organization, int64, error) {
	limit, offset, err := normalizePage(limit, offset)

	if err != nil {
		return nil, 0, utils.ClientError{Err: err}
	}

	organizations, err := s.admin.ListOrganizations(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.admin.CountOrganizations(ctx)
	return organizations, total, err
}

func (s *AdminService) Tunnels(ctx context.Context, limit, offset int) ([]models.Tunnel, int64, error) {
	limit, offset, err := normalizePage(limit, offset)

	if err != nil {
		return nil, 0, utils.ClientError{Err: err}
	}

	tunnels, err := s.admin.ListTunnels(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.admin.CountTunnels(ctx)
	return tunnels, total, err
}

func (s *AdminService) Subscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, int64, error) {
	limit, offset, err := normalizePage(limit, offset)

	if err != nil {
		return nil, 0, utils.ClientError{Err: err}
	}

	subscriptions, err := s.admin.ListSubscriptions(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.admin.CountSubscriptions(ctx)
	return subscriptions, total, err
}

func (s *AdminService) SetUserStatus(ctx context.Context, userID string, status models.UserStatus) error {

	if status != models.UserStatusActive && status != models.UserStatusDisabled {
		return fmt.Errorf("unsupported user status")
	}

	return s.admin.SetUserStatus(ctx, userID, status)
}

func (s *AdminService) AuditLogs(ctx context.Context, limit, offset int) ([]models.AuditEvent, int64, error) {
	limit, offset, err := normalizePage(limit, offset)

	if err != nil {
		return nil, 0, utils.ClientError{Err: err}
	}

	events, err := s.admin.ListAuditEvents(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.admin.CountAuditEvents(ctx)
	return events, total, err
}

func (s *AdminService) Usage(ctx context.Context) (repositories.AdminUsage, error) {
	return s.admin.Usage(ctx)
}

func normalizePage(limit, offset int) (int, int, error) {

	if limit == 0 {
		limit = 50
	}

	if limit < 1 || limit > 100 || offset < 0 || offset > 100_000 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 100 and offset must be between 0 and 100000")
	}

	return limit, offset, nil
}

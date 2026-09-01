package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type AdminRepository interface {
	PlatformAdminRepository
	FindUser(context.Context, string) (models.User, error)
	FindOrganization(context.Context, string) (models.Organization, error)
	ListUsers(context.Context, int, int) ([]models.User, error)
	CountUsers(context.Context) (int64, error)
	SetUserStatus(context.Context, string, models.UserStatus) error
	ListOrganizations(context.Context, int, int) ([]models.Organization, error)
	CountOrganizations(context.Context) (int64, error)
	ListTunnels(context.Context, int, int) ([]models.Tunnel, error)
	CountTunnels(context.Context) (int64, error)
	ListSubscriptions(context.Context, int, int) ([]models.Subscription, error)
	CountSubscriptions(context.Context) (int64, error)
	ListAuditEvents(context.Context, int, int) ([]models.AuditEvent, error)
	CountAuditEvents(context.Context) (int64, error)
	Usage(context.Context) (AdminUsage, error)
}

type PlatformAdminRepository interface {
	CountPlatformAdmins(context.Context) (int64, error)
	CreatePlatformAdmin(context.Context, *models.PlatformAdmin) error
	IsPlatformAdmin(context.Context, string) (bool, error)
}

type AdminUsage struct {
	BandwidthBytes int64 `json:"bandwidthBytes"`
	RequestCount   int64 `json:"requestCount"`
	ErrorCount     int64 `json:"errorCount"`
}

type GormAdminRepository struct{ db *gorm.DB }

func (r *GormAdminRepository) FindUser(ctx context.Context, id string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil { return models.User{}, mapError(err) }
	return user, nil
}

func (r *GormAdminRepository) FindOrganization(ctx context.Context, id string) (models.Organization, error) {
	var organization models.Organization
	if err := r.db.WithContext(ctx).First(&organization, "id = ?", id).Error; err != nil { return models.Organization{}, mapError(err) }
	return organization, nil
}

func NewAdminRepository(db *gorm.DB) (*GormAdminRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormAdminRepository{db: db}, nil
}

func (r *GormAdminRepository) CountPlatformAdmins(ctx context.Context) (int64, error) {
	return r.count(ctx, &models.PlatformAdmin{}, "active = true", "count platform admins")
}

func (r *GormAdminRepository) CreatePlatformAdmin(ctx context.Context, admin *models.PlatformAdmin) error {

	if admin == nil || admin.UserID == "" {
		return fmt.Errorf("platform admin user is required")
	}

	return wrap(r.db.WithContext(ctx).Create(admin).Error, "create platform admin")
}

func (r *GormAdminRepository) IsPlatformAdmin(ctx context.Context, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PlatformAdmin{}).Where("user_id = ? AND active = true", userID).Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("check platform admin: %w", err)
	}

	return count == 1, nil
}

func (r *GormAdminRepository) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("created_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error
	return users, wrap(err, "list admin users")
}

func (r *GormAdminRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.count(ctx, &models.User{}, "deleted_at IS NULL", "count users")
}

func (r *GormAdminRepository) SetUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ? AND deleted_at IS NULL", userID).Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("set user status: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormAdminRepository) ListOrganizations(ctx context.Context, limit, offset int) ([]models.Organization, error) {
	var organizations []models.Organization
	err := r.db.WithContext(ctx).Order("created_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&organizations).Error
	return organizations, wrap(err, "list admin organizations")
}

func (r *GormAdminRepository) CountOrganizations(ctx context.Context) (int64, error) {
	return r.count(ctx, &models.Organization{}, "1 = 1", "count organizations")
}

func (r *GormAdminRepository) ListTunnels(ctx context.Context, limit, offset int) ([]models.Tunnel, error) {
	var tunnels []models.Tunnel
	err := r.db.WithContext(ctx).Order("created_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&tunnels).Error
	return tunnels, wrap(err, "list admin tunnels")
}

func (r *GormAdminRepository) CountTunnels(ctx context.Context) (int64, error) {
	return r.count(ctx, &models.Tunnel{}, "1 = 1", "count tunnels")
}

func (r *GormAdminRepository) ListSubscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	err := r.db.WithContext(ctx).Order("created_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error
	return subscriptions, wrap(err, "list admin subscriptions")
}

func (r *GormAdminRepository) CountSubscriptions(ctx context.Context) (int64, error) {
	return r.count(ctx, &models.Subscription{}, "1 = 1", "count subscriptions")
}

func (r *GormAdminRepository) ListAuditEvents(ctx context.Context, limit, offset int) ([]models.AuditEvent, error) {
	var events []models.AuditEvent
	err := r.db.WithContext(ctx).Order("occurred_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&events).Error
	return events, wrap(err, "list admin audit events")
}

func (r *GormAdminRepository) CountAuditEvents(ctx context.Context) (int64, error) {
	return r.count(ctx, &models.AuditEvent{}, "1 = 1", "count audit events")
}

func (r *GormAdminRepository) Usage(ctx context.Context) (AdminUsage, error) {
	var usage AdminUsage
	err := r.db.WithContext(ctx).Model(&models.UsageEvent{}).Select("COALESCE(SUM(bytes), 0) AS bandwidth_bytes, COALESCE(SUM(CASE WHEN event_type = 'request' THEN 1 ELSE 0 END), 0) AS request_count, COALESCE(SUM(CASE WHEN event_type = 'error' OR status_code >= 400 THEN 1 ELSE 0 END), 0) AS error_count").Scan(&usage).Error
	return usage, wrap(err, "aggregate admin usage")
}

func (r *GormAdminRepository) count(ctx context.Context, model any, query, operation string) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(model).Where(query).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	return count, nil
}

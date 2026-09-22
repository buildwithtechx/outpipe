package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type UserRepository interface {
	FindByID(context.Context, string) (models.User, error)
	FindByEmail(context.Context, string) (models.User, error)
	Create(context.Context, *models.User) error
	UpdateLastLogin(context.Context, string, time.Time) error
	Delete(context.Context, string, time.Time) error
}

type OAuthIdentityRepository interface {
	Find(context.Context, string, string) (models.OAuthIdentity, error)
	Save(context.Context, *models.OAuthIdentity) error
}

type SessionRepository interface {
	Create(context.Context, *models.Session) error
	FindActive(context.Context, string, time.Time) (models.Session, error)
	Touch(context.Context, string, time.Time) error
	Revoke(context.Context, string, time.Time) error
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type APIKeyRepository interface {
	Create(context.Context, *models.APIKey) error
	FindByPrefix(context.Context, string) (models.APIKey, error)
	FindByID(context.Context, string) (models.APIKey, error)
	ListByUser(context.Context, string, *string) ([]models.APIKey, error)
	Touch(context.Context, string, time.Time) error
	Revoke(context.Context, string, time.Time) error
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type DeviceLoginRepository interface {
	Create(context.Context, *models.DeviceLogin) error
	FindPending(context.Context, string, time.Time) (models.DeviceLogin, error)
	Complete(context.Context, string, string, string, time.Time) error
	StoreToken(context.Context, string, string) error
	FindAuthenticated(context.Context, string, time.Time) (models.DeviceLogin, error)
	ConsumeToken(context.Context, string, time.Time) (models.DeviceLogin, error)
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type GormUserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) (*GormUserRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormUserRepository{db: db}, nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		return models.User{}, mapError(err)
	}

	return user, nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).Where("lower(email) = ? AND deleted_at IS NULL", strings.ToLower(strings.TrimSpace(email))).First(&user).Error; err != nil {
		return models.User{}, mapError(err)
	}

	return user, nil
}

func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {

	if user == nil {
		return fmt.Errorf("user is required")
	}

	return wrap(r.db.WithContext(ctx).Create(user).Error, "create user")
}

func (r *GormUserRepository) UpdateLastLogin(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(map[string]any{"last_login_at": at}).Error, "update last login")
}

func (r *GormUserRepository) Delete(ctx context.Context, id string, at time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", at)

	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

type GormOAuthIdentityRepository struct{ db *gorm.DB }

func NewOAuthIdentityRepository(db *gorm.DB) (*GormOAuthIdentityRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormOAuthIdentityRepository{db: db}, nil
}

func (r *GormOAuthIdentityRepository) Find(ctx context.Context, provider string, subject string) (models.OAuthIdentity, error) {
	var identity models.OAuthIdentity

	if err := r.db.WithContext(ctx).Where("provider = ? AND subject = ?", provider, subject).First(&identity).Error; err != nil {
		return models.OAuthIdentity{}, mapError(err)
	}

	return identity, nil
}

func (r *GormOAuthIdentityRepository) Save(ctx context.Context, identity *models.OAuthIdentity) error {

	if identity == nil {
		return fmt.Errorf("oauth identity is required")
	}

	return wrap(r.db.WithContext(ctx).Save(identity).Error, "save oauth identity")
}

func mapError(err error) error {

	if err == nil {
		return nil
	}

	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

func wrap(err error, operation string) error {

	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", operation, err)
}

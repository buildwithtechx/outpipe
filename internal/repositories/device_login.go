package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type GormDeviceLoginRepository struct{ db *gorm.DB }

func NewDeviceLoginRepository(db *gorm.DB) (*GormDeviceLoginRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormDeviceLoginRepository{db: db}, nil
}

func (r *GormDeviceLoginRepository) Create(ctx context.Context, login *models.DeviceLogin) error {

	if login == nil {
		return fmt.Errorf("device login is required")
	}

	return wrap(r.db.WithContext(ctx).Create(login).Error, "create device login")
}

func (r *GormDeviceLoginRepository) FindPending(ctx context.Context, codeHash string, now time.Time) (models.DeviceLogin, error) {
	var login models.DeviceLogin
	err := r.db.WithContext(ctx).Where("code_hash = ? AND status = ? AND expires_at > ?", codeHash, "pending", now).First(&login).Error

	if err != nil {
		return models.DeviceLogin{}, mapError(err)
	}

	return login, nil
}

func (r *GormDeviceLoginRepository) Complete(ctx context.Context, id string, userID string, tokenHash string, at time.Time) error {
	result := r.db.WithContext(ctx).Model(&models.DeviceLogin{}).Where("id = ? AND status = ?", id, "pending").Updates(map[string]any{
		"user_id": userID, "user_token_hash": tokenHash, "status": "authenticated", "completed_at": at,
	})

	if result.Error != nil {
		return fmt.Errorf("complete device login: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *GormDeviceLoginRepository) StoreToken(ctx context.Context, id, token string) error {
	return wrap(r.db.WithContext(ctx).Model(&models.DeviceLogin{}).Where("id = ?", id).Update("user_token", token).Error, "store device token")
}

func (r *GormDeviceLoginRepository) FindAuthenticated(ctx context.Context, codeHash string, now time.Time) (models.DeviceLogin, error) {
	var login models.DeviceLogin

	if err := r.db.WithContext(ctx).Where("code_hash = ? AND status = ? AND expires_at > ?", codeHash, "authenticated", now).First(&login).Error; err != nil {
		return models.DeviceLogin{}, mapError(err)
	}

	return login, nil
}

func (r *GormDeviceLoginRepository) ConsumeToken(ctx context.Context, tokenHash string, now time.Time) (models.DeviceLogin, error) {
	var login models.DeviceLogin
	err := r.db.WithContext(ctx).Where("user_token_hash = ? AND status = ? AND expires_at > ?", tokenHash, "authenticated", now).First(&login).Error

	if err != nil {
		return models.DeviceLogin{}, mapError(err)
	}

	result := r.db.WithContext(ctx).Model(&models.DeviceLogin{}).Where("id = ? AND status = ?", login.ID, "authenticated").Updates(map[string]any{"status": "consumed"})

	if result.Error != nil {
		return models.DeviceLogin{}, fmt.Errorf("consume device login: %w", result.Error)
	}

	if result.RowsAffected != 1 {
		return models.DeviceLogin{}, ErrNotFound
	}

	return login, nil
}

func (r *GormDeviceLoginRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at <= ? OR status IN ?", now, []string{"expired", "consumed"}).Delete(&models.DeviceLogin{})
	return result.RowsAffected, wrap(result.Error, "delete expired device logins")
}

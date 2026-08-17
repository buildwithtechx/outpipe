package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type GormSessionRepository struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) (*GormSessionRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormSessionRepository{db: db}, nil
}

func (r *GormSessionRepository) Create(ctx context.Context, session *models.Session) error {
	if session == nil {
		return fmt.Errorf("session is required")
	}
	return wrap(r.db.WithContext(ctx).Create(session).Error, "create session")
}

func (r *GormSessionRepository) FindActive(ctx context.Context, tokenHash string, now time.Time) (models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, now).First(&session).Error
	if err != nil {
		return models.Session{}, mapError(err)
	}
	return session, nil
}

func (r *GormSessionRepository) Touch(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.Session{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"last_seen_at": at}).Error, "touch session")
}

func (r *GormSessionRepository) Revoke(ctx context.Context, id string, at time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.Session{}).Where("id = ? AND revoked_at IS NULL", id).Updates(map[string]any{"revoked_at": at}).Error, "revoke session")
}

func (r *GormSessionRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at <= ? OR revoked_at IS NOT NULL", now).Delete(&models.Session{})
	return result.RowsAffected, wrap(result.Error, "delete expired sessions")
}

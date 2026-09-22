package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

type EmailRepository interface {
	Create(context.Context, *models.EmailDelivery) error
	ClaimPending(context.Context, time.Time, int) ([]models.EmailDelivery, error)
	MarkSent(context.Context, string, time.Time) error
	MarkFailed(context.Context, string, string, time.Time) error
}

type GormEmailRepository struct{ db *gorm.DB }

func NewEmailRepository(db *gorm.DB) (*GormEmailRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormEmailRepository{db: db}, nil
}

func (r *GormEmailRepository) Create(ctx context.Context, email *models.EmailDelivery) error {
	if email == nil {
		return fmt.Errorf("email delivery is required")
	}
	return wrap(r.db.WithContext(ctx).Create(email).Error, "create email delivery")
}

func (r *GormEmailRepository) ClaimPending(ctx context.Context, now time.Time, limit int) ([]models.EmailDelivery, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("email delivery limit must be between 1 and 100")
	}
	var deliveries []models.EmailDelivery
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reclaimBefore := now.Add(-10 * time.Minute)
		if err := tx.Where("((status IN ? AND available_at <= ?) OR (status = ? AND updated_at < ?))", []models.EmailDeliveryStatus{models.EmailDeliveryPending, models.EmailDeliveryFailed}, now, models.EmailDeliverySending, reclaimBefore).
			Order("available_at ASC").Limit(limit).Clauses(clauseLockSkipLocked()).Find(&deliveries).Error; err != nil {
			return err
		}
		if len(deliveries) == 0 {
			return nil
		}
		ids := make([]string, 0, len(deliveries))
		for i := range deliveries {
			deliveries[i].Status = models.EmailDeliverySending
			deliveries[i].Attempts++
			ids = append(ids, deliveries[i].ID)
		}
		return tx.Model(&models.EmailDelivery{}).Where("id IN ?", ids).Updates(map[string]any{"status": models.EmailDeliverySending, "attempts": gorm.Expr("attempts + 1")}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("claim email deliveries: %w", err)
	}
	return deliveries, nil
}

func (r *GormEmailRepository) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.EmailDelivery{}).Where("id = ?", id).Updates(map[string]any{"status": models.EmailDeliverySent, "sent_at": sentAt, "last_error": ""}).Error, "mark email sent")
}

func (r *GormEmailRepository) MarkFailed(ctx context.Context, id, message string, availableAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.EmailDelivery{}).Where("id = ?", id).Updates(map[string]any{"status": models.EmailDeliveryFailed, "last_error": message, "available_at": availableAt}).Error, "mark email failed")
}

func clauseLockSkipLocked() clause.Expression {
	return clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}
}

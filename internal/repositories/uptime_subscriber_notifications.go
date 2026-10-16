package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func (r *GormUptimeRepository) CreateSubscriberNotification(ctx context.Context, notification *models.UptimeSubscriberNotification) error {
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return fmt.Errorf("create subscriber notification: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) ClaimPendingSubscriberNotifications(ctx context.Context, now time.Time, limit int) ([]models.UptimeSubscriberNotification, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("subscriber notification limit must be between 1 and 100")
	}
	var notifications []models.UptimeSubscriberNotification
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reclaimBefore := now.Add(-10 * time.Minute)
		if err := tx.Where("((status IN ? AND available_at <= ?) OR (status = ? AND updated_at < ?))", []models.SubscriberNotificationStatus{models.SubscriberNotificationPending, models.SubscriberNotificationFailed}, now, models.SubscriberNotificationSending, reclaimBefore).
			Order("available_at ASC").Limit(limit).Clauses(clauseLockSkipLocked()).Find(&notifications).Error; err != nil {
			return err
		}
		if len(notifications) == 0 {
			return nil
		}
		ids := make([]string, 0, len(notifications))
		for i := range notifications {
			notifications[i].Status = models.SubscriberNotificationSending
			notifications[i].Attempts++
			ids = append(ids, notifications[i].ID)
		}
		return tx.Model(&models.UptimeSubscriberNotification{}).Where("id IN ?", ids).Updates(map[string]any{"status": models.SubscriberNotificationSending, "attempts": gorm.Expr("attempts + 1")}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("claim subscriber notifications: %w", err)
	}
	return notifications, nil
}

func (r *GormUptimeRepository) MarkSubscriberNotificationSent(ctx context.Context, id string, sentAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).Where("id = ?", id).Updates(map[string]any{"status": models.SubscriberNotificationSent, "sent_at": sentAt, "last_error": ""}).Error, "mark subscriber notification sent")
}

func (r *GormUptimeRepository) MarkSubscriberNotificationFailed(ctx context.Context, id, message string, availableAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).Where("id = ?", id).Updates(map[string]any{"status": models.SubscriberNotificationFailed, "last_error": message, "available_at": availableAt}).Error, "mark subscriber notification failed")
}

func (r *GormUptimeRepository) MarkSubscriberNotificationDeadLetter(ctx context.Context, id, message string) error {
	return wrap(r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).Where("id = ?", id).Updates(map[string]any{"status": models.SubscriberNotificationDeadLetter, "last_error": message}).Error, "mark subscriber notification dead letter")
}

func (r *GormUptimeRepository) RequeueSubscriberNotification(ctx context.Context, id string, availableAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).Where("id = ?", id).Updates(map[string]any{"status": models.SubscriberNotificationPending, "attempts": gorm.Expr("attempts - 1"), "available_at": availableAt}).Error, "requeue subscriber notification")
}

func (r *GormUptimeRepository) CountSubscriberNotificationsByStatus(ctx context.Context, pageID string) (map[models.SubscriberNotificationStatus]int64, *time.Time, error) {
	counts := map[models.SubscriberNotificationStatus]int64{
		models.SubscriberNotificationPending:    0,
		models.SubscriberNotificationSending:    0,
		models.SubscriberNotificationSent:       0,
		models.SubscriberNotificationFailed:     0,
		models.SubscriberNotificationDeadLetter: 0,
	}
	type statusCount struct {
		Status models.SubscriberNotificationStatus
		Total  int64
	}
	var rows []statusCount
	if err := r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).
		Joins("JOIN uptime_subscribers ON uptime_subscribers.id = uptime_subscriber_notifications.subscriber_id").
		Where("uptime_subscribers.status_page_id = ?", pageID).
		Select("uptime_subscriber_notifications.status as status, count(*) as total").
		Group("uptime_subscriber_notifications.status").
		Scan(&rows).Error; err != nil {
		return nil, nil, fmt.Errorf("count subscriber notifications: %w", err)
	}
	for _, row := range rows {
		counts[row.Status] = row.Total
	}
	var oldest *time.Time
	var oldestTime sql.NullTime
	activeStatuses := []models.SubscriberNotificationStatus{
		models.SubscriberNotificationPending,
		models.SubscriberNotificationFailed,
		models.SubscriberNotificationSending,
	}
	if err := r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).
		Joins("JOIN uptime_subscribers ON uptime_subscribers.id = uptime_subscriber_notifications.subscriber_id").
		Where("uptime_subscribers.status_page_id = ? AND uptime_subscriber_notifications.status IN ?", pageID, activeStatuses).
		Select("MIN(uptime_subscriber_notifications.created_at)").
		Scan(&oldestTime).Error; err != nil {
		return nil, nil, fmt.Errorf("find oldest subscriber notification: %w", err)
	} else if oldestTime.Valid && !oldestTime.Time.IsZero() {
		oldest = &oldestTime.Time
	}
	return counts, oldest, nil
}

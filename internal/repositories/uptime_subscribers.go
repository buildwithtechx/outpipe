package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func (r *GormUptimeRepository) GetStatusPageByID(ctx context.Context, id string) (*models.UptimeStatusPage, error) {
	var page models.UptimeStatusPage
	if err := r.db.WithContext(ctx).First(&page, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("status page %q not found: %w", id, err)
		}
		return nil, fmt.Errorf("find status page %q: %w", id, err)
	}
	return &page, nil
}

func (r *GormUptimeRepository) SetSubscriptionsEnabled(ctx context.Context, pageID string, enabled bool) (*models.UptimeStatusPage, error) {
	page, err := r.GetStatusPageByID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	page.SubscriptionsEnabled = enabled
	page.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(page).Error; err != nil {
		return nil, fmt.Errorf("update subscriber settings: %w", err)
	}
	return page, nil
}

func (r *GormUptimeRepository) GetSubscriber(ctx context.Context, id string) (*models.UptimeSubscriber, error) {
	var subscriber models.UptimeSubscriber
	if err := r.db.WithContext(ctx).First(&subscriber, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscriber %q not found: %w", id, err)
		}
		return nil, fmt.Errorf("find subscriber %q: %w", id, err)
	}
	return &subscriber, nil
}

func (r *GormUptimeRepository) GetSubscriberByEmail(ctx context.Context, pageID, email string) (*models.UptimeSubscriber, error) {
	var subscriber models.UptimeSubscriber
	result := r.db.WithContext(ctx).Where("status_page_id = ? AND email = ?", pageID, email).Limit(1).Find(&subscriber)
	if result.Error != nil {
		return nil, fmt.Errorf("find subscriber: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("subscriber not found: %w", gorm.ErrRecordNotFound)
	}
	return &subscriber, nil
}

func (r *GormUptimeRepository) GetSubscriberByTokenHash(ctx context.Context, tokenHash string) (*models.UptimeSubscriber, error) {
	var subscriber models.UptimeSubscriber
	result := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Limit(1).Find(&subscriber)
	if result.Error != nil {
		return nil, fmt.Errorf("find subscriber by token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("subscriber not found: %w", gorm.ErrRecordNotFound)
	}
	return &subscriber, nil
}

func (r *GormUptimeRepository) SaveSubscriber(ctx context.Context, subscriber *models.UptimeSubscriber) error {
	if err := r.db.WithContext(ctx).Save(subscriber).Error; err != nil {
		return fmt.Errorf("save subscriber: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) DeleteSubscriber(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.UptimeSubscriberComponent{}, "subscriber_id = ?", id).Error; err != nil {
		return fmt.Errorf("delete subscriber components: %w", err)
	}
	if err := r.db.WithContext(ctx).Delete(&models.UptimeSubscriber{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete subscriber: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) ListSubscribersFiltered(ctx context.Context, pageID, query string, status models.SubscriberStatus, limit int, before time.Time, beforeID string) ([]models.UptimeSubscriber, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	db := r.db.WithContext(ctx).Where("status_page_id = ?", pageID)
	if status != "" {
		db = db.Where("status = ?", string(status))
	}
	trimmed := strings.TrimSpace(query)
	if trimmed != "" {
		pattern := "%" + escapeLikePattern(strings.ToLower(trimmed)) + "%"
		db = db.Where("LOWER(email) LIKE ? ESCAPE '\\'", pattern)
	}
	if !before.IsZero() {
		db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", before, before, beforeID)
	}
	var subscribers []models.UptimeSubscriber
	if err := db.Order("created_at DESC").Order("id DESC").Limit(limit + 1).Find(&subscribers).Error; err != nil {
		return nil, fmt.Errorf("list subscribers: %w", err)
	}
	return subscribers, nil
}

func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return replacer.Replace(value)
}

func (r *GormUptimeRepository) CountSubscribersByStatus(ctx context.Context, pageID string) (map[models.SubscriberStatus]int64, error) {
	type statusCount struct {
		Status string
		Total  int64
	}
	var rows []statusCount
	if err := r.db.WithContext(ctx).Model(&models.UptimeSubscriber{}).Select("status, COUNT(*) AS total").Where("status_page_id = ?", pageID).Group("status").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("count subscribers: %w", err)
	}
	counts := map[models.SubscriberStatus]int64{
		models.SubscriberStatusPending:      0,
		models.SubscriberStatusConfirmed:    0,
		models.SubscriberStatusUnsubscribed: 0,
	}
	for _, row := range rows {
		counts[models.SubscriberStatus(row.Status)] = row.Total
	}
	return counts, nil
}

func (r *GormUptimeRepository) SetSubscriberComponents(ctx context.Context, subscriberID string, monitorIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.UptimeSubscriberComponent{}, "subscriber_id = ?", subscriberID).Error; err != nil {
			return fmt.Errorf("clear subscriber components: %w", err)
		}
		for _, monitorID := range monitorIDs {
			trimmed := strings.TrimSpace(monitorID)
			if trimmed == "" {
				continue
			}
			if err := tx.Create(&models.UptimeSubscriberComponent{SubscriberID: subscriberID, MonitorID: trimmed}).Error; err != nil {
				return fmt.Errorf("add subscriber component: %w", err)
			}
		}
		return nil
	})
}

func (r *GormUptimeRepository) ListSubscriberComponents(ctx context.Context, subscriberID string) ([]string, error) {
	var rows []models.UptimeSubscriberComponent
	if err := r.db.WithContext(ctx).Where("subscriber_id = ?", subscriberID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list subscriber components: %w", err)
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.MonitorID)
	}
	return ids, nil
}

func (r *GormUptimeRepository) ListNotificationTargets(ctx context.Context, pageID string, monitorIDs []string) ([]models.UptimeSubscriber, error) {
	var subscribers []models.UptimeSubscriber
	db := r.db.WithContext(ctx).Where("status_page_id = ? AND status = ?", pageID, string(models.SubscriberStatusConfirmed))
	if len(monitorIDs) > 0 {
		db = db.Where("NOT EXISTS (SELECT 1 FROM uptime_subscriber_components AS scope WHERE scope.subscriber_id = uptime_subscribers.id) OR EXISTS (SELECT 1 FROM uptime_subscriber_components AS wanted WHERE wanted.subscriber_id = uptime_subscribers.id AND wanted.monitor_id IN ?)", monitorIDs)
	}
	if err := db.Order("created_at ASC").Limit(1000).Find(&subscribers).Error; err != nil {
		return nil, fmt.Errorf("list notification targets: %w", err)
	}
	return subscribers, nil
}

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

func (r *GormUptimeRepository) RequeueSubscriberNotification(ctx context.Context, id string, availableAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).Where("id = ?", id).Updates(map[string]any{"status": models.SubscriberNotificationPending, "attempts": gorm.Expr("attempts - 1"), "available_at": availableAt}).Error, "requeue subscriber notification")
}

func (r *GormUptimeRepository) MarkSubscriberSent(ctx context.Context, subscriberID string, sentAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.UptimeSubscriber{}).Where("id = ?", subscriberID).Updates(map[string]any{"last_sent_at": sentAt}).Error, "mark subscriber sent")
}

func (r *GormUptimeRepository) CountSubscriberSendsSince(ctx context.Context, subscriberID string, since time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UptimeSubscriberNotification{}).Where("subscriber_id = ? AND status = ? AND sent_at >= ?", subscriberID, string(models.SubscriberNotificationSent), since).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count subscriber sends: %w", err)
	}
	return count, nil
}

func (r *GormUptimeRepository) RecordSubscriptionAttempt(ctx context.Context, attempt *models.UptimeSubscriptionAttempt) error {
	if err := r.db.WithContext(ctx).Create(attempt).Error; err != nil {
		return fmt.Errorf("record subscription attempt: %w", err)
	}
	return nil
}

func (r *GormUptimeRepository) CountRecentAttempts(ctx context.Context, pageID, ipHash, emailHash string, since time.Time) (int64, int64, error) {
	var byIP int64
	if err := r.db.WithContext(ctx).Model(&models.UptimeSubscriptionAttempt{}).Where("status_page_id = ? AND ip_hash = ? AND created_at >= ?", pageID, ipHash, since).Count(&byIP).Error; err != nil {
		return 0, 0, fmt.Errorf("count subscription attempts: %w", err)
	}
	var byEmail int64
	if err := r.db.WithContext(ctx).Model(&models.UptimeSubscriptionAttempt{}).Where("status_page_id = ? AND email_hash = ? AND created_at >= ?", pageID, emailHash, since).Count(&byEmail).Error; err != nil {
		return 0, 0, fmt.Errorf("count subscription attempts: %w", err)
	}
	return byIP, byEmail, nil
}

func (r *GormUptimeRepository) PruneSubscriptionAttempts(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&models.UptimeSubscriptionAttempt{})
	return result.RowsAffected, wrap(result.Error, "prune subscription attempts")
}

func (r *GormUptimeRepository) DeleteExpiredPendingSubscribers(ctx context.Context, pageID string, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("status_page_id = ? AND status = ? AND token_expires_at IS NOT NULL AND token_expires_at < ?", pageID, string(models.SubscriberStatusPending), now).Delete(&models.UptimeSubscriber{})
	return result.RowsAffected, wrap(result.Error, "prune expired pending subscribers")
}

func (r *GormUptimeRepository) GetIncidentUpdate(ctx context.Context, id string) (*models.UptimeIncidentUpdate, error) {
	var update models.UptimeIncidentUpdate
	if err := r.db.WithContext(ctx).First(&update, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("incident update %q not found: %w", id, err)
		}
		return nil, fmt.Errorf("find incident update %q: %w", id, err)
	}
	return &update, nil
}

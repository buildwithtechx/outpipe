package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func (s *UptimeService) ProcessPendingSubscriberNotifications(ctx context.Context, limit int) error {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	notifications, err := s.repo.ClaimPendingSubscriberNotifications(ctx, time.Now().UTC(), limit)
	if err != nil {
		return err
	}
	for i := range notifications {
		notification := notifications[i]
		if err := s.sendSubscriberNotification(ctx, &notification); err != nil {
			now := time.Now().UTC()
			if notification.Attempts >= subscriberNotifyMaxTries {
				if markErr := s.repo.MarkSubscriberNotificationDeadLetter(ctx, notification.ID, err.Error()); markErr != nil {
					return fmt.Errorf("mark subscriber notification %s dead-letter: %w", notification.ID, markErr)
				}
			} else {
				availableAt := now.Add(webhookRetryDelay(notification.Attempts))
				if markErr := s.repo.MarkSubscriberNotificationFailed(ctx, notification.ID, err.Error(), availableAt); markErr != nil {
					return fmt.Errorf("mark subscriber notification %s failed: %w", notification.ID, markErr)
				}
			}
		}
	}
	return nil
}

func (s *UptimeService) sendSubscriberNotification(ctx context.Context, notification *models.UptimeSubscriberNotification) error {
	now := time.Now().UTC()
	drop := func() error {
		if err := s.repo.MarkSubscriberNotificationSent(ctx, notification.ID, now); err != nil {
			return fmt.Errorf("drop stale subscriber notification %s: %w", notification.ID, err)
		}
		return nil
	}
	isNotFound := func(err error) bool {
		return errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, repositories.ErrNotFound)
	}

	subscriber, err := s.repo.GetSubscriber(ctx, notification.SubscriberID)
	if err != nil {
		if isNotFound(err) {
			return drop()
		}
		return err
	}
	if subscriber.Status != models.SubscriberStatusConfirmed {
		return drop()
	}
	update, err := s.repo.GetIncidentUpdate(ctx, notification.UpdateID)
	if err != nil {
		if isNotFound(err) {
			return drop()
		}
		return err
	}
	incident, err := s.repo.GetIncident(ctx, update.IncidentID)
	if err != nil {
		if isNotFound(err) {
			return drop()
		}
		return err
	}
	page, err := s.repo.GetStatusPageByOrg(ctx, incident.OrganizationID)
	if err != nil {
		if isNotFound(err) {
			return drop()
		}
		return err
	}
	if !page.Published || !page.SubscriptionsEnabled {
		return drop()
	}
	sent, err := s.repo.CountSubscriberSendsSince(ctx, subscriber.ID, now.Add(-24*time.Hour))
	if err != nil {
		return err
	}
	if sent >= maxSendsPerSubscriberDay {
		if err := s.repo.RequeueSubscriberNotification(ctx, notification.ID, now.Add(24*time.Hour)); err != nil {
			return fmt.Errorf("requeue capped subscriber notification %s: %w", notification.ID, err)
		}
		return nil
	}
	if s.subscriberMailer == nil {
		return fmt.Errorf("subscriber mailer is required")
	}
	unsubscribeURL := s.UnsubscribeLinkFor(subscriber.ID, page.Slug)
	if err := s.subscriberMailer.SendSubscriberIncident(ctx, subscriber.Email, page.Title, incident.Title, string(update.Status), update.Message, unsubscribeURL); err != nil {
		return err
	}
	if err := s.repo.MarkSubscriberNotificationSent(ctx, notification.ID, now); err != nil {
		return fmt.Errorf("mark subscriber notification %s sent: %w", notification.ID, err)
	}
	if err := s.repo.MarkSubscriberSent(ctx, subscriber.ID, now); err != nil {
		return fmt.Errorf("mark subscriber %s sent: %w", subscriber.ID, err)
	}
	return nil
}

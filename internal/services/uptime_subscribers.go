package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

var (
	ErrSubscriptionsDisabled = errors.New("subscriptions are disabled for this status page")
	ErrConfirmTokenInvalid   = errors.New("confirmation link is invalid or expired")
	ErrUnsubscribeInvalid    = errors.New("unsubscribe link is invalid or expired")
)

const (
	subscriberConfirmTTL     = 24 * time.Hour
	subscriberUnsubscribeTTL = 365 * 24 * time.Hour
	maxConfirmedSubscribers  = 1000
	maxPendingSubscribers    = 2000
	maxSignupsPerIPPerHour   = 5
	maxSignupsPerEmailPer15m = 1
	maxSendsPerSubscriberDay = 20
	subscriberNotifyMaxTries = 5
)

type SubscriberMailer interface {
	SendSubscriberConfirm(ctx context.Context, email, pageTitle, confirmURL string) error
	SendSubscriberIncident(ctx context.Context, email, pageTitle, incidentTitle, status, message, unsubscribeURL string) error
}

func (s *UptimeService) SetSubscriberMailer(mailer SubscriberMailer) {
	s.subscriberMailer = mailer
}

func (s *UptimeService) SetSubscriberSigningKey(key string) {
	s.subscriberSigningKey = key
}

func (s *UptimeService) SetSubscriberBaseURL(baseURL string) {
	s.subscriberBaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

func (s *UptimeService) signingKey() string {
	if s.subscriberSigningKey != "" {
		return s.subscriberSigningKey
	}
	return "outpipe-subscriber-fallback"
}

func (s *UptimeService) Subscribe(ctx context.Context, slug, email string) error {
	_, _, err := s.RequestSubscription(ctx, slug, email, nil, "")
	return err
}

func (s *UptimeService) RequestSubscription(ctx context.Context, slug, email string, componentIDs []string, clientIP string) (string, string, error) {
	page, err := s.repo.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		return "", "", fmt.Errorf("status page %q not found: %w", slug, err)
	}
	if !page.Published {
		return "", "", repositories.ErrNotFound
	}
	if !page.SubscriptionsEnabled {
		return "", "", ErrSubscriptionsDisabled
	}
	normalized, err := normalizeSubscriberEmail(email)
	if err != nil {
		return "", "", err
	}
	now := time.Now().UTC()
	_, _ = s.repo.DeleteExpiredPendingSubscribers(ctx, page.ID, now)
	_, _ = s.repo.PruneSubscriptionAttempts(ctx, now.Add(-48*time.Hour))
	scoped, err := s.validateSubscriberScope(ctx, page.OrganizationID, componentIDs)
	if err != nil {
		return "", "", err
	}
	ipHash := hashRateLimitKey(s.signingKey(), "ip", strings.TrimSpace(clientIP))
	emailHash := hashRateLimitKey(s.signingKey(), "email", page.ID+":"+normalized)
	byIP, _, err := s.repo.CountRecentAttempts(ctx, page.ID, ipHash, emailHash, now.Add(-time.Hour))
	if err != nil {
		return "", "", err
	}
	_, byEmail, err := s.repo.CountRecentAttempts(ctx, page.ID, ipHash, emailHash, now.Add(-15*time.Minute))
	if err != nil {
		return "", "", err
	}
	counts, err := s.repo.CountSubscribersByStatus(ctx, page.ID)
	if err != nil {
		return "", "", err
	}
	limited := byIP >= maxSignupsPerIPPerHour || byEmail >= maxSignupsPerEmailPer15m
	full := counts[models.SubscriberStatusConfirmed] >= maxConfirmedSubscribers || counts[models.SubscriberStatusPending] >= maxPendingSubscribers
	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		return "", "", fmt.Errorf("generate attempt id: %w", err)
	}
	_ = s.repo.RecordSubscriptionAttempt(ctx, &models.UptimeSubscriptionAttempt{ID: hex.EncodeToString(randomID), StatusPageID: page.ID, IPHash: ipHash, EmailHash: emailHash, CreatedAt: now})
	if limited || full {
		return "", "", nil
	}
	existing, err := s.repo.GetSubscriberByEmail(ctx, page.ID, normalized)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", err
	}
	if existing != nil && existing.Status == models.SubscriberStatusConfirmed {
		return "", "", nil
	}
	subscriberID := ""
	if existing != nil {
		subscriberID = existing.ID
	} else {
		rawID := make([]byte, 16)
		if _, err := rand.Read(rawID); err != nil {
			return "", "", fmt.Errorf("generate subscriber id: %w", err)
		}
		subscriberID = hex.EncodeToString(rawID)
	}
	expiresAt := now.Add(subscriberConfirmTTL)
	token, err := mintSubscriberToken(s.signingKey(), subscriberID, "confirm", expiresAt)
	if err != nil {
		return "", "", err
	}
	subscriber := models.UptimeSubscriber{
		ID:             subscriberID,
		StatusPageID:   page.ID,
		Email:          normalized,
		Status:         models.SubscriberStatusPending,
		Confirmed:      false,
		ConfirmedAt:    nil,
		TokenHash:      hashSubscriberToken(token),
		TokenExpiresAt: &expiresAt,
		UnsubscribedAt: nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if existing != nil {
		subscriber.CreatedAt = existing.CreatedAt
		subscriber.LastSentAt = existing.LastSentAt
	}
	if err := s.repo.SaveSubscriber(ctx, &subscriber); err != nil {
		return "", "", err
	}
	if err := s.repo.SetSubscriberComponents(ctx, subscriberID, scoped); err != nil {
		return "", "", err
	}
	if s.subscriberMailer != nil {
		confirmURL := s.subscriberConfirmURL(page.Slug, token)
		if err := s.subscriberMailer.SendSubscriberConfirm(ctx, normalized, page.Title, confirmURL); err != nil {
			return "", "", err
		}
	}
	return subscriberID, token, nil
}

func (s *UptimeService) ConfirmSubscription(ctx context.Context, slug, token string) (models.UptimeSubscriber, error) {
	verified, err := verifySubscriberToken(s.signingKey(), strings.TrimSpace(token), "confirm")
	if err != nil {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	now := time.Now().UTC()
	if now.After(verified.expiresAt) {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	subscriber, err := s.repo.GetSubscriberByTokenHash(ctx, hashSubscriberToken(strings.TrimSpace(token)))
	if err != nil {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	if subscriber.ID != verified.subscriberID || subscriber.Status != models.SubscriberStatusPending {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	if subscriber.TokenExpiresAt == nil || now.After(*subscriber.TokenExpiresAt) {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	page, err := s.repo.GetStatusPageByID(ctx, subscriber.StatusPageID)
	if err != nil {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	if slug != "" && !strings.EqualFold(strings.TrimSpace(slug), page.Slug) {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	if !page.Published || !page.SubscriptionsEnabled {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	counts, err := s.repo.CountSubscribersByStatus(ctx, page.ID)
	if err != nil {
		return models.UptimeSubscriber{}, err
	}
	if counts[models.SubscriberStatusConfirmed] >= maxConfirmedSubscribers {
		return models.UptimeSubscriber{}, ErrConfirmTokenInvalid
	}
	subscriber.Status = models.SubscriberStatusConfirmed
	subscriber.Confirmed = true
	subscriber.ConfirmedAt = &now
	subscriber.TokenHash = ""
	subscriber.TokenExpiresAt = nil
	subscriber.UpdatedAt = now
	if err := s.repo.SaveSubscriber(ctx, subscriber); err != nil {
		return models.UptimeSubscriber{}, err
	}
	return *subscriber, nil
}

func (s *UptimeService) UnsubscribeWithToken(ctx context.Context, slug, token string) (models.UptimeSubscriber, error) {
	verified, err := verifySubscriberToken(s.signingKey(), strings.TrimSpace(token), "unsubscribe")
	if err != nil {
		return models.UptimeSubscriber{}, ErrUnsubscribeInvalid
	}
	if time.Now().UTC().After(verified.expiresAt) {
		return models.UptimeSubscriber{}, ErrUnsubscribeInvalid
	}
	subscriber, err := s.repo.GetSubscriber(ctx, verified.subscriberID)
	if err != nil {
		return models.UptimeSubscriber{}, ErrUnsubscribeInvalid
	}
	page, err := s.repo.GetStatusPageByID(ctx, subscriber.StatusPageID)
	if err != nil {
		return models.UptimeSubscriber{}, ErrUnsubscribeInvalid
	}
	if slug != "" && !strings.EqualFold(strings.TrimSpace(slug), page.Slug) {
		return models.UptimeSubscriber{}, ErrUnsubscribeInvalid
	}
	return s.markUnsubscribed(ctx, subscriber)
}

func (s *UptimeService) markUnsubscribed(ctx context.Context, subscriber *models.UptimeSubscriber) (models.UptimeSubscriber, error) {
	now := time.Now().UTC()
	subscriber.Status = models.SubscriberStatusUnsubscribed
	subscriber.Confirmed = false
	subscriber.TokenHash = ""
	subscriber.TokenExpiresAt = nil
	subscriber.UnsubscribedAt = &now
	subscriber.UpdatedAt = now
	if err := s.repo.SaveSubscriber(ctx, subscriber); err != nil {
		return models.UptimeSubscriber{}, err
	}
	return *subscriber, nil
}

func (s *UptimeService) UnsubscribeLinkFor(subscriberID, slug string) string {
	token := mintUnsubscribeToken(s.signingKey(), subscriberID, time.Now().UTC().Add(subscriberUnsubscribeTTL))
	base := s.subscriberBaseURL
	if base == "" {
		base = "https://status.outpipe.dev"
	}
	return base + "/status/" + slug + "/unsubscribe?token=" + token
}

func (s *UptimeService) subscriberConfirmURL(slug, token string) string {
	base := s.subscriberBaseURL
	if base == "" {
		base = "https://status.outpipe.dev"
	}
	return base + "/status/" + slug + "/confirm?token=" + token
}

func (s *UptimeService) validateSubscriberScope(ctx context.Context, orgID string, componentIDs []string) ([]string, error) {
	trimmed := make([]string, 0, len(componentIDs))
	for _, id := range componentIDs {
		if value := strings.TrimSpace(id); value != "" {
			trimmed = append(trimmed, value)
		}
	}
	if len(trimmed) == 0 {
		return nil, nil
	}
	if len(trimmed) > 100 {
		return nil, fmt.Errorf("at most 100 components can be selected")
	}
	monitors, err := s.repo.ListOrgMonitors(ctx, orgID)
	if err != nil {
		return nil, err
	}
	owned := make(map[string]struct{}, len(monitors))
	for _, monitor := range monitors {
		owned[monitor.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(trimmed))
	scoped := make([]string, 0, len(trimmed))
	for _, id := range trimmed {
		if _, ok := owned[id]; !ok {
			return nil, fmt.Errorf("component %q is not part of this status page", id)
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		scoped = append(scoped, id)
	}
	return scoped, nil
}

func (s *UptimeService) NotifyIncidentUpdate(ctx context.Context, incidentID, updateID string, monitorIDs []string) (int, error) {
	incident, err := s.repo.GetIncident(ctx, incidentID)
	if err != nil {
		return 0, err
	}
	page, err := s.repo.GetStatusPageByOrg(ctx, incident.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if !page.Published || !page.SubscriptionsEnabled {
		return 0, nil
	}
	targets, err := s.repo.ListNotificationTargets(ctx, page.ID, monitorIDs)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	enqueued := 0
	for _, target := range targets {
		rawID := make([]byte, 16)
		if _, err := rand.Read(rawID); err != nil {
			return enqueued, fmt.Errorf("generate notification id: %w", err)
		}
		notification := models.UptimeSubscriberNotification{
			ID:           hex.EncodeToString(rawID),
			SubscriberID: target.ID,
			UpdateID:     updateID,
			Status:       models.SubscriberNotificationPending,
			AvailableAt:  now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.repo.CreateSubscriberNotification(ctx, &notification); err != nil {
			continue
		}
		enqueued++
	}
	return enqueued, nil
}

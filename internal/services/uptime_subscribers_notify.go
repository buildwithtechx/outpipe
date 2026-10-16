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
	"outpipe.dev/outpipe/pkg/utils"
)

func (s *UptimeService) UnsubscribeLinkFor(subscriberID, slug string) string {
	token := mintUnsubscribeToken(s.signingKey(), subscriberID, time.Now().UTC().Add(subscriberUnsubscribeTTL))
	base := utils.FirstNonEmpty(s.subscriberBaseURL, "https://status.outpipe.dev")
	return base + "/status/" + slug + "/unsubscribe?token=" + token
}

func (s *UptimeService) subscriberConfirmURL(slug, token string) string {
	base := utils.FirstNonEmpty(s.subscriberBaseURL, "https://status.outpipe.dev")
	return base + "/status/" + slug + "/confirm?token=" + token
}

func (s *UptimeService) validateSubscriberScope(ctx context.Context, orgID string, componentIDs []string) ([]string, error) {
	trimmed := utils.CompactStrings(componentIDs)
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
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") || strings.Contains(msg, "primary") {
				continue
			}
			return enqueued, fmt.Errorf("create subscriber notification for %s: %w", target.ID, err)
		}
		enqueued++
	}
	return enqueued, nil
}

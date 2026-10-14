package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type SubscriberView struct {
	ID             string                  `json:"id"`
	StatusPageID   string                  `json:"statusPageId"`
	Email          string                  `json:"email"`
	Status         models.SubscriberStatus `json:"status"`
	Confirmed      bool                    `json:"confirmed"`
	ConfirmedAt    *time.Time              `json:"confirmedAt,omitempty"`
	UnsubscribedAt *time.Time              `json:"unsubscribedAt,omitempty"`
	LastSentAt     *time.Time              `json:"lastSentAt,omitempty"`
	CreatedAt      time.Time               `json:"createdAt"`
	Components     []string                `json:"components"`
}

type SubscriberListResult struct {
	Subscribers          []SubscriberView `json:"subscribers"`
	Counts               map[string]int64 `json:"counts"`
	NextCursor           string           `json:"nextCursor,omitempty"`
	SubscriptionsEnabled bool             `json:"subscriptionsEnabled"`
}

func (s *UptimeService) GetStatusPageByID(ctx context.Context, pageID string) (*models.UptimeStatusPage, error) {
	return s.repo.GetStatusPageByID(ctx, strings.TrimSpace(pageID))
}

func (s *UptimeService) StatusPageOrganizationID(ctx context.Context, pageID string) (string, error) {
	page, err := s.repo.GetStatusPageByID(ctx, strings.TrimSpace(pageID))
	if err != nil {
		return "", err
	}
	return page.OrganizationID, nil
}

func (s *UptimeService) ListManagedSubscribers(ctx context.Context, pageID, query, status string, limit int, cursor string) (SubscriberListResult, error) {
	var result SubscriberListResult
	page, err := s.repo.GetStatusPageByID(ctx, strings.TrimSpace(pageID))
	if err != nil {
		return result, err
	}
	filter := models.SubscriberStatus(strings.ToLower(strings.TrimSpace(status)))
	if status != "" && strings.ToLower(strings.TrimSpace(status)) != "all" {
		switch filter {
		case models.SubscriberStatusPending, models.SubscriberStatusConfirmed, models.SubscriberStatusUnsubscribed:
		default:
			return result, fmt.Errorf("unknown subscriber status %q", status)
		}
	} else {
		filter = ""
	}
	if len(query) > 254 {
		return result, fmt.Errorf("search query is too long")
	}
	before, beforeID, err := decodeSubscriberCursor(cursor)
	if err != nil {
		return result, err
	}
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.repo.ListSubscribersFiltered(ctx, page.ID, query, filter, limit, before, beforeID)
	if err != nil {
		return result, err
	}
	counts, err := s.repo.CountSubscribersByStatus(ctx, page.ID)
	if err != nil {
		return result, err
	}
	views := make([]SubscriberView, 0, len(rows))
	for i, row := range rows {
		if i == limit {
			last := rows[limit-1]
			result.NextCursor = encodeSubscriberCursor(last.CreatedAt, last.ID)
			break
		}
		components, err := s.repo.ListSubscriberComponents(ctx, row.ID)
		if err != nil {
			return result, err
		}
		views = append(views, SubscriberView{
			ID:             row.ID,
			StatusPageID:   row.StatusPageID,
			Email:          row.Email,
			Status:         row.Status,
			Confirmed:      row.Confirmed,
			ConfirmedAt:    row.ConfirmedAt,
			UnsubscribedAt: row.UnsubscribedAt,
			LastSentAt:     row.LastSentAt,
			CreatedAt:      row.CreatedAt,
			Components:     components,
		})
	}
	result.Subscribers = views
	result.SubscriptionsEnabled = page.SubscriptionsEnabled
	result.Counts = map[string]int64{
		"pending":      counts[models.SubscriberStatusPending],
		"confirmed":    counts[models.SubscriberStatusConfirmed],
		"unsubscribed": counts[models.SubscriberStatusUnsubscribed],
	}
	return result, nil
}

func (s *UptimeService) ManagerUnsubscribe(ctx context.Context, pageID, subscriberID string) (SubscriberView, error) {
	var view SubscriberView
	subscriber, err := s.ownedSubscriber(ctx, pageID, subscriberID)
	if err != nil {
		return view, err
	}
	updated, err := s.markUnsubscribed(ctx, subscriber)
	if err != nil {
		return view, err
	}
	return s.toSubscriberView(ctx, &updated)
}

func (s *UptimeService) RemoveSubscriber(ctx context.Context, pageID, subscriberID string) error {
	if _, err := s.ownedSubscriber(ctx, pageID, subscriberID); err != nil {
		return err
	}
	return s.repo.DeleteSubscriber(ctx, strings.TrimSpace(subscriberID))
}

func (s *UptimeService) UpdateSubscriberSettings(ctx context.Context, pageID string, enabled bool) (*models.UptimeStatusPage, error) {
	if _, err := s.repo.GetStatusPageByID(ctx, strings.TrimSpace(pageID)); err != nil {
		return nil, err
	}
	return s.repo.SetSubscriptionsEnabled(ctx, strings.TrimSpace(pageID), enabled)
}

func (s *UptimeService) SetSubscriberScope(ctx context.Context, pageID, subscriberID string, componentIDs []string) (SubscriberView, error) {
	var view SubscriberView
	subscriber, err := s.ownedSubscriber(ctx, pageID, subscriberID)
	if err != nil {
		return view, err
	}
	page, err := s.repo.GetStatusPageByID(ctx, subscriber.StatusPageID)
	if err != nil {
		return view, err
	}
	scoped, err := s.validateSubscriberScope(ctx, page.OrganizationID, componentIDs)
	if err != nil {
		return view, err
	}
	if err := s.repo.SetSubscriberComponents(ctx, subscriber.ID, scoped); err != nil {
		return view, err
	}
	return s.toSubscriberView(ctx, subscriber)
}

func (s *UptimeService) ownedSubscriber(ctx context.Context, pageID, subscriberID string) (*models.UptimeSubscriber, error) {
	subscriber, err := s.repo.GetSubscriber(ctx, strings.TrimSpace(subscriberID))
	if err != nil {
		return nil, err
	}
	if subscriber.StatusPageID != strings.TrimSpace(pageID) {
		return nil, fmt.Errorf("subscriber not found: %w", gorm.ErrRecordNotFound)
	}
	if _, err := s.repo.GetStatusPageByID(ctx, strings.TrimSpace(pageID)); err != nil {
		return nil, err
	}
	return subscriber, nil
}

func (s *UptimeService) toSubscriberView(ctx context.Context, subscriber *models.UptimeSubscriber) (SubscriberView, error) {
	var view SubscriberView
	if subscriber == nil {
		return view, fmt.Errorf("subscriber is required")
	}
	components, err := s.repo.ListSubscriberComponents(ctx, subscriber.ID)
	if err != nil {
		return view, err
	}
	return SubscriberView{
		ID:             subscriber.ID,
		StatusPageID:   subscriber.StatusPageID,
		Email:          subscriber.Email,
		Status:         subscriber.Status,
		Confirmed:      subscriber.Confirmed,
		ConfirmedAt:    subscriber.ConfirmedAt,
		UnsubscribedAt: subscriber.UnsubscribedAt,
		LastSentAt:     subscriber.LastSentAt,
		CreatedAt:      subscriber.CreatedAt,
		Components:     components,
	}, nil
}

func (s *UptimeService) AuthorizeStatusPage(ctx context.Context, pageID, orgID string) error {
	page, err := s.repo.GetStatusPageByID(ctx, strings.TrimSpace(pageID))
	if err != nil {
		return err
	}
	if page.OrganizationID != orgID {
		return repositories.ErrNotFound
	}
	return nil
}

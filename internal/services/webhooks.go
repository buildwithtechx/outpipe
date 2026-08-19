package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/security"
)

const webhookDeliveryTimeout = 5 * time.Second

type WebhookService struct {
	subscriptions repositories.WebhookRepository
	client        *http.Client
	now           func() time.Time
}

func NewWebhookService(subscriptions repositories.WebhookRepository) (*WebhookService, error) {

	if subscriptions == nil {
		return nil, fmt.Errorf("webhook repository is required")
	}

	return &WebhookService{
		subscriptions: subscriptions,
		client:        &http.Client{Timeout: webhookDeliveryTimeout},
		now:           time.Now,
	}, nil
}

func (s *WebhookService) Create(ctx context.Context, organizationID, name, webhookURL string, events []string) (models.WebhookSubscription, string, error) {

	if organizationID == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(webhookURL) == "" {
		return models.WebhookSubscription{}, "", fmt.Errorf("organization, name, and url are required")
	}

	parsed, err := url.ParseRequestURI(strings.TrimSpace(webhookURL))

	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return models.WebhookSubscription{}, "", fmt.Errorf("webhook url must be a valid http or https url")
	}

	for _, event := range events {

		if !models.ValidWebhookEvent(event) {
			return models.WebhookSubscription{}, "", fmt.Errorf("invalid webhook event %q", event)
		}
	}

	if len(events) == 0 {
		events = []string{string(models.WebhookEventTunnelConnected), string(models.WebhookEventTunnelDisconnected), string(models.WebhookEventTunnelRevoked)}
	}

	eventJSON, err := json.Marshal(events)

	if err != nil {
		return models.WebhookSubscription{}, "", fmt.Errorf("encode webhook events: %w", err)
	}

	secret, err := randomSecret()

	if err != nil {
		return models.WebhookSubscription{}, "", err
	}

	subscription := models.WebhookSubscription{OrganizationID: organizationID, Name: strings.TrimSpace(name), URL: parsed.String(), Events: string(eventJSON), Secret: secret}

	if err := s.subscriptions.Create(ctx, &subscription); err != nil {
		return models.WebhookSubscription{}, "", fmt.Errorf("create webhook subscription: %w", err)
	}

	return subscription, secret, nil
}

func (s *WebhookService) List(ctx context.Context, organizationID string) ([]models.WebhookSubscription, error) {
	return s.subscriptions.ListByOrganization(ctx, organizationID)
}

func (s *WebhookService) Delete(ctx context.Context, organizationID, subscriptionID string) error {

	subscription, err := s.subscriptions.FindByID(ctx, subscriptionID)

	if err != nil {
		return fmt.Errorf("find webhook subscription: %w", err)
	}

	if subscription.OrganizationID != organizationID {
		return fmt.Errorf("webhook subscription does not belong to this organization")
	}

	if err := s.subscriptions.Delete(ctx, subscriptionID); err != nil {
		return fmt.Errorf("delete webhook subscription: %w", err)
	}

	return nil
}

func (s *WebhookService) Deliveries(ctx context.Context, organizationID, subscriptionID string) ([]models.WebhookDelivery, error) {

	subscription, err := s.subscriptions.FindByID(ctx, subscriptionID)

	if err != nil {
		return nil, fmt.Errorf("find webhook subscription: %w", err)
	}

	if subscription.OrganizationID != organizationID {
		return nil, fmt.Errorf("webhook subscription does not belong to this organization")
	}

	return s.subscriptions.ListDeliveries(ctx, subscriptionID)
}

func (s *WebhookService) Dispatch(ctx context.Context, organizationID string, event models.WebhookEvent, data map[string]any) {

	subscriptions, err := s.subscriptions.ListByOrganization(ctx, organizationID)

	if err != nil {
		return
	}

	for _, subscription := range subscriptions {

		if !subscriptionReceives(subscription.Events, string(event)) {
			continue
		}

		payload := s.payloadFor(event, data)
		s.deliver(ctx, &subscription, payload)
	}
}

func (s *WebhookService) payloadFor(event models.WebhookEvent, data map[string]any) []byte {
	envelope := map[string]any{
		"id":         newEventID(),
		"type":       string(event),
		"occurredAt": s.now().UTC().Format(time.RFC3339),
		"data":       data,
	}
	payload, err := json.Marshal(envelope)

	if err != nil {
		return []byte("{}")
	}

	return payload
}

func newEventID() string {
	return uuid.New().String()
}

func (s *WebhookService) deliver(ctx context.Context, subscription *models.WebhookSubscription, payload []byte) {
	eventType, eventID := envelopeIdentity(payload)
	delivery := models.WebhookDelivery{SubscriptionID: subscription.ID, EventID: eventID, EventType: eventType, Payload: string(payload), Attempts: 1, Status: "delivered"}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, subscription.URL, bytes.NewReader(payload))

	if err != nil {
		delivery.Status = "failed"
		delivery.Error = fmt.Sprintf("build delivery request: %v", err)
		_ = s.subscriptions.CreateDelivery(ctx, &delivery)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Outpipe-Event-ID", eventID)
	request.Header.Set("X-Outpipe-Event-Type", eventType)
	request.Header.Set("X-Outpipe-Signature", security.SignHMACSHA256(payload, subscription.Secret))

	response, err := s.client.Do(request)

	if err != nil {
		delivery.Status = "failed"
		delivery.Error = fmt.Sprintf("deliver webhook: %v", err)
		_ = s.subscriptions.CreateDelivery(ctx, &delivery)
		return
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		delivery.Status = "failed"
		delivery.Error = fmt.Sprintf("webhook receiver returned status %d", response.StatusCode)
		_ = s.subscriptions.CreateDelivery(ctx, &delivery)
		return
	}

	now := s.now()
	delivery.DeliveredAt = &now
	_ = s.subscriptions.CreateDelivery(ctx, &delivery)
	subscription.LastDeliveredAt = &now
}

func envelopeIdentity(payload []byte) (eventType, eventID string) {
	var envelope struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal(payload, &envelope); err != nil {
		return "", ""
	}

	return envelope.Type, envelope.ID
}

func subscriptionReceives(eventsJSON, event string) bool {
	var events []string

	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil {
		return false
	}

	return slices.Contains(events, event)
}

func randomSecret() (string, error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}

	return hex.EncodeToString(raw), nil
}

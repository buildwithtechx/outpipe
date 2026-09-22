package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/infra/telemetry"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/validation"
)

const webhookDeliveryTimeout = 5 * time.Second
const webhookMaxDeliveryAttempts = 5

type WebhookService struct {
	subscriptions repositories.WebhookRepository
	client        *http.Client
	now           func() time.Time
	validateURL   func(string) error
	synchronous   bool
	metrics       *telemetry.MetricsExporter
}

type WebhookServiceOption func(*WebhookService)

func WithWebhookHTTPClient(client *http.Client) WebhookServiceOption {
	return func(service *WebhookService) {
		if client != nil {
			service.client = client
		}
	}
}

func WithWebhookURLValidator(validateURL func(string) error) WebhookServiceOption {
	return func(service *WebhookService) {
		if validateURL != nil {
			service.validateURL = validateURL
		}
	}
}

// WithWebhookSynchronousDelivery is intended for integration tests only. The
// production service queues deliveries for the cron worker.
func WithWebhookSynchronousDelivery() WebhookServiceOption {
	return func(service *WebhookService) { service.synchronous = true }
}

func WithWebhookMetrics(metrics *telemetry.MetricsExporter) WebhookServiceOption {
	return func(service *WebhookService) { service.metrics = metrics }
}

func NewWebhookService(subscriptions repositories.WebhookRepository, options ...WebhookServiceOption) (*WebhookService, error) {

	if subscriptions == nil {
		return nil, fmt.Errorf("webhook repository is required")
	}

	service := &WebhookService{
		subscriptions: subscriptions,
		client:        validation.NewSafeHTTPClient(webhookDeliveryTimeout),
		now:           time.Now,
		validateURL:   validation.ValidateWebhookURL,
	}

	for _, option := range options {
		option(service)
	}

	return service, nil
}

func (s *WebhookService) Create(ctx context.Context, organizationID, name, webhookURL string, events []string) (models.WebhookSubscription, string, error) {

	if organizationID == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(webhookURL) == "" {
		return models.WebhookSubscription{}, "", fmt.Errorf("organization, name, and url are required")
	}

	if err := s.validateURL(webhookURL); err != nil {
		return models.WebhookSubscription{}, "", err
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

	subscription := models.WebhookSubscription{OrganizationID: organizationID, Name: strings.TrimSpace(name), URL: strings.TrimSpace(webhookURL), Events: string(eventJSON), Secret: secret}

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

func (s *WebhookService) prepareDeliveries(ctx context.Context, organizationID string, event models.WebhookEvent, data map[string]any) ([]models.WebhookDelivery, error) {
	subscriptions, err := s.subscriptions.ListByOrganization(ctx, organizationID)

	if err != nil {
		return nil, fmt.Errorf("list webhook subscriptions: %w", err)
	}

	payload := s.payloadFor(event, data)
	deliveries := make([]models.WebhookDelivery, 0, len(subscriptions))
	for _, subscription := range subscriptions {

		if !subscriptionReceives(subscription.Events, string(event)) {
			continue
		}

		deliveries = append(deliveries, models.WebhookDelivery{
			SubscriptionID: subscription.ID,
			EventID:        eventID(payload),
			EventType:      string(event),
			Payload:        string(payload),
			Status:         models.WebhookDeliveryPending,
			AvailableAt:    s.now().UTC(),
		})
	}

	return deliveries, nil
}

func (s *WebhookService) afterEnqueue(ctx context.Context, count int) {
	if s.metrics != nil && count > 0 {
		s.metrics.IncCounter("outpipe_webhook_deliveries_enqueued_total", int64(count))
	}

	if s.synchronous && count > 0 {
		_ = s.ProcessPending(ctx, count)
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

func (s *WebhookService) ProcessPending(ctx context.Context, limit int) error {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	deliveries, err := s.subscriptions.ClaimPendingDeliveries(ctx, s.now().UTC(), limit)
	if err != nil {
		return err
	}
	if s.metrics != nil {
		if queued, countErr := s.subscriptions.CountQueuedDeliveries(ctx); countErr == nil {
			s.metrics.SetGauge("outpipe_webhook_queue_depth", queued)
		}
	}

	for i := range deliveries {
		if s.metrics != nil {
			s.metrics.IncCounter("outpipe_webhook_delivery_attempts_total", 1)
		}
		if err := s.deliver(ctx, &deliveries[i]); err != nil {
			if deliveries[i].Attempts >= webhookMaxDeliveryAttempts {
				deliveries[i].Status = models.WebhookDeliveryFailed
				deliveries[i].AvailableAt = s.now().UTC()
			} else {
				deliveries[i].Status = models.WebhookDeliveryFailed
				deliveries[i].AvailableAt = s.now().UTC().Add(webhookRetryDelay(deliveries[i].Attempts))
			}
			deliveries[i].Error = err.Error()
			if updateErr := s.subscriptions.UpdateDelivery(ctx, &deliveries[i]); updateErr != nil {
				return fmt.Errorf("mark webhook delivery %s failed: %w", deliveries[i].ID, updateErr)
			}
			if s.metrics != nil {
				s.metrics.IncCounter("outpipe_webhook_deliveries_failed_total", 1)
				if deliveries[i].Attempts >= webhookMaxDeliveryAttempts {
					s.metrics.IncCounter("outpipe_webhook_deliveries_dead_lettered_total", 1)
				} else {
					s.metrics.IncCounter("outpipe_webhook_delivery_retries_total", 1)
				}
			}
			continue
		}

		now := s.now().UTC()
		deliveries[i].Status = models.WebhookDeliveryDelivered
		deliveries[i].DeliveredAt = &now
		deliveries[i].Error = ""
		if updateErr := s.subscriptions.UpdateDelivery(ctx, &deliveries[i]); updateErr != nil {
			return fmt.Errorf("mark webhook delivery %s sent: %w", deliveries[i].ID, updateErr)
		}
		if s.metrics != nil {
			s.metrics.IncCounter("outpipe_webhook_deliveries_sent_total", 1)
		}
		if deliveries[i].Subscription != nil {
			deliveries[i].Subscription.LastDeliveredAt = &now
			if updateErr := s.subscriptions.Update(ctx, deliveries[i].Subscription); updateErr != nil {
				return fmt.Errorf("update webhook subscription %s: %w", deliveries[i].Subscription.ID, updateErr)
			}
		}
	}

	return nil
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

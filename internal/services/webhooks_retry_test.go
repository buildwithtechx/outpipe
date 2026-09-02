package services

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

type webhookRetryRepository struct {
	delivery *models.WebhookDelivery
	updated  []models.WebhookDelivery
}

func (r *webhookRetryRepository) Create(context.Context, *models.WebhookSubscription) error {
	return nil
}

func (r *webhookRetryRepository) Update(context.Context, *models.WebhookSubscription) error {
	return nil
}

func (r *webhookRetryRepository) FindByID(context.Context, string) (models.WebhookSubscription, error) {
	if r.delivery != nil && r.delivery.Subscription != nil {
		return *r.delivery.Subscription, nil
	}
	return models.WebhookSubscription{}, errors.New("subscription not found")
}

func (r *webhookRetryRepository) ListByOrganization(context.Context, string) ([]models.WebhookSubscription, error) {
	return nil, nil
}

func (r *webhookRetryRepository) Delete(context.Context, string) error { return nil }

func (r *webhookRetryRepository) CreateDelivery(_ context.Context, delivery *models.WebhookDelivery) error {
	r.delivery = delivery
	return nil
}

func (r *webhookRetryRepository) UpdateDelivery(_ context.Context, delivery *models.WebhookDelivery) error {
	r.updated = append(r.updated, *delivery)
	return nil
}

func (r *webhookRetryRepository) ClaimPendingDeliveries(_ context.Context, _ time.Time, _ int) ([]models.WebhookDelivery, error) {
	if r.delivery == nil {
		return nil, nil
	}
	delivery := *r.delivery
	delivery.Attempts++
	return []models.WebhookDelivery{delivery}, nil
}

func (r *webhookRetryRepository) CountQueuedDeliveries(context.Context) (int64, error) { return 1, nil }

func (r *webhookRetryRepository) ListDeliveries(context.Context, string) ([]models.WebhookDelivery, error) {
	return r.updated, nil
}

type failingWebhookTransport struct{}

func (failingWebhookTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("receiver timed out")
}

func TestWebhookRetrySchedulesBackoffAndDeadLetters(t *testing.T) {
	repository := &webhookRetryRepository{delivery: &models.WebhookDelivery{
		Base:         models.Base{ID: "delivery-1"},
		EventID:      "event-1",
		EventType:    string(models.WebhookEventTunnelConnected),
		Payload:      `{"id":"event-1","type":"tunnel.connected"}`,
		Status:       models.WebhookDeliveryPending,
		Subscription: &models.WebhookSubscription{Base: models.Base{ID: "subscription-1"}, URL: "https://hooks.example.test/events", Secret: "secret"},
	}}
	service, err := NewWebhookService(repository, WithWebhookHTTPClient(&http.Client{Transport: failingWebhookTransport{}}))
	if err != nil {
		t.Fatal(err)
	}

	if err := service.ProcessPending(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if len(repository.updated) != 1 || repository.updated[0].Status != models.WebhookDeliveryFailed {
		t.Fatalf("first failure = %#v", repository.updated)
	}
	if got := time.Until(repository.updated[0].AvailableAt); got < 0 {
		t.Fatalf("retry should be scheduled in the future, got %s", got)
	}

	repository.delivery.Attempts = webhookMaxDeliveryAttempts - 1
	if err := service.ProcessPending(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	last := repository.updated[len(repository.updated)-1]
	if last.Status != models.WebhookDeliveryFailed || last.Attempts != webhookMaxDeliveryAttempts {
		t.Fatalf("dead-letter transition = %#v", last)
	}
}

package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/security"
)

func webhookRetryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	if attempts > 6 {
		attempts = 6
	}
	return time.Duration(1<<uint(attempts-1)) * time.Minute
}

func (s *WebhookService) deliver(ctx context.Context, delivery *models.WebhookDelivery) error {
	if delivery.Subscription == nil {
		return fmt.Errorf("webhook subscription is missing")
	}

	subscription := delivery.Subscription
	payload := []byte(delivery.Payload)
	eventType, eventID := envelopeIdentity(payload)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, subscription.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build delivery request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Outpipe-Event-ID", eventID)
	request.Header.Set("X-Outpipe-Event-Type", eventType)
	request.Header.Set("X-Outpipe-Signature", security.SignHMACSHA256(payload, subscription.Secret))

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("deliver webhook: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("webhook receiver returned status %d", response.StatusCode)
	}

	return nil
}

func eventID(payload []byte) string {
	_, id := envelopeIdentity(payload)
	return id
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

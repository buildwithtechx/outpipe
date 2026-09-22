package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

func parseWebhook(provider string, payload []byte) (services.BillingTransition, string, string, error) {
	var envelope struct {
		ID    string         `json:"id"`
		Event string         `json:"type"`
		Data  map[string]any `json:"data"`
		Type  string         `json:"event"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return services.BillingTransition{}, "", "", fmt.Errorf("decode billing webhook: %w", err)
	}

	eventType := envelope.Event
	if eventType == "" {
		eventType = envelope.Type
	}
	data := envelope.Data
	if nested, ok := data["subscription"].(map[string]any); ok {
		for key, value := range nested {
			if _, exists := data[key]; !exists {
				data[key] = value
			}
		}
	}

	attemptKeys := []string{"attempts_remaining", "attemptsRemaining", "retries_remaining"}
	transition := services.BillingTransition{
		Provider:             models.BillingProvider(strings.ToLower(provider)),
		ProviderSubscription: stringValue(data, "subscription_id", "subscription_code", "id"),
		ProviderCustomer:     stringValue(data, "customer_id", "customer_code"),
		ProviderProduct:      stringValue(data, "product_id"),
		ProviderInvoice:      stringValue(data, "invoice_id", "order_id", "invoice_code", "reference"),
		InvoiceURL:           stringValue(data, "invoice_url", "receipt_url", "order_url"),
		Status:               subscriptionStatus(eventType, stringValue(data, "status")),
		CurrentPeriodEnd:     timeValue(data, "current_period_end", "period_end"),
		CancelAtPeriodEnd:    boolValue(data, "cancel_at_period_end", "cancelled"),
		EventType:            eventType,
		AmountMinor:          int64Value(data, "amount", "amount_minor", "amountMinor"),
		Currency:             stringValue(data, "currency"),
		AttemptsRemaining:    intValue(data, attemptKeys...),
		AttemptsKnown:        hasIntValue(data, attemptKeys...),
		PreviousPlan:         stringValue(data, "previous_plan", "previousPlan"),
		PaidAt:               timeValue(data, "paid_at", "paidAt", "created_at"),
	}

	if metadata, ok := data["metadata"].(map[string]any); ok {
		if transition.ProviderSubscription == "" {
			transition.ProviderSubscription = stringValue(metadata, "subscription_id")
		}
		transition.BillingInterval = stringValue(metadata, "billing_interval")
	}
	if transition.BillingInterval == "" {
		transition.BillingInterval = stringValue(data, "billing_interval")
	}
	if provider == string(models.BillingProviderPaystack) {
		transition.ProviderAuthorization = stringValue(data, "authorization_code")
	}

	return transition, envelope.ID, eventType, nil
}

func subscriptionStatus(eventType, value string) models.SubscriptionStatus {
	value = strings.ToLower(value)
	if value == "active" || value == "trialing" || value == "past_due" || value == "paused" || value == "canceled" || value == "expired" {
		return models.SubscriptionStatus(value)
	}
	switch {
	case strings.Contains(value, "success"), strings.Contains(eventType, "activate"), strings.Contains(eventType, "paid"):
		return models.SubscriptionStatusActive
	case strings.Contains(eventType, "cancel"):
		return models.SubscriptionStatusCanceled
	case strings.Contains(eventType, "pause"):
		return models.SubscriptionStatusPaused
	case strings.Contains(eventType, "fail"), strings.Contains(eventType, "past_due"):
		return models.SubscriptionStatusPastDue
	default:
		return ""
	}
}

func stringValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func boolValue(values map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := values[key].(bool); ok {
			return value
		}
	}
	return false
}

func intValue(values map[string]any, keys ...string) int {
	for _, key := range keys {
		switch value := values[key].(type) {
		case float64:
			return int(value)
		case int:
			return value
		case string:
			if parsed, err := strconv.Atoi(value); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func hasIntValue(values map[string]any, keys ...string) bool {
	for _, key := range keys {
		switch value := values[key].(type) {
		case float64, int:
			return true
		case string:
			if _, err := strconv.Atoi(value); err == nil {
				return true
			}
		}
	}
	return false
}

func int64Value(values map[string]any, keys ...string) int64 {
	for _, key := range keys {
		switch value := values[key].(type) {
		case float64:
			return int64(value)
		case int:
			return int64(value)
		case int64:
			return value
		case string:
			if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func timeValue(values map[string]any, keys ...string) *time.Time {
	for _, key := range keys {
		if value, ok := values[key].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return &parsed
			}
		}
		if value, ok := values[key].(float64); ok {
			parsed := time.Unix(int64(value), 0).UTC()
			return &parsed
		}
	}
	return nil
}

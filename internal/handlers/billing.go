package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/infra/billing"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

type BillingHandler struct {
	billing       *services.BillingService
	alerts        *services.AlertService
	webhookSecret string
	polarSecret   string
	paystack      *billing.PaystackClient
}

type CheckoutRequest struct {
	PlanKey         string `json:"planKey" validate:"required"`
	BillingInterval string `json:"billingInterval" validate:"omitempty,oneof=month year"`
}

func NewBillingHandler(billing *services.BillingService) (*BillingHandler, error) {

	if billing == nil {
		return nil, fmt.Errorf("billing service is required")
	}

	return &BillingHandler{billing: billing}, nil
}

func (h *BillingHandler) SetWebhookSecret(secret string) { h.webhookSecret = secret }
func (h *BillingHandler) SetProviderSecrets(polarSecret string, paystack *billing.PaystackClient) {
	h.polarSecret = polarSecret
	h.paystack = paystack
}
func (h *BillingHandler) SetAlerts(alerts *services.AlertService) { h.alerts = alerts }

func (h *BillingHandler) Status(c *fiber.Ctx) error {
	plan, subscription, err := h.billing.Entitlements(c.UserContext(), c.Params("organizationID"))

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(fiber.Map{"plan": plan, "subscription": subscription})
}

func (h *BillingHandler) Checkout(c *fiber.Ctx) error {
	var input CheckoutRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	url, err := h.billing.Checkout(c.UserContext(), c.Params("organizationID"), input.PlanKey, input.BillingInterval)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(fiber.Map{"url": url})
}

func (h *BillingHandler) Portal(c *fiber.Ctx) error {
	url, err := h.billing.Portal(c.UserContext(), c.Params("organizationID"))

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(fiber.Map{"url": url})
}

func (h *BillingHandler) Cancel(c *fiber.Ctx) error {

	if err := h.billing.Cancel(c.UserContext(), c.Params("organizationID")); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *BillingHandler) Resume(c *fiber.Ctx) error {

	if err := h.billing.Resume(c.UserContext(), c.Params("organizationID")); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *BillingHandler) Invoices(c *fiber.Ctx) error {
	invoices, err := h.billing.ListInvoices(c.UserContext(), c.Params("organizationID"))

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(fiber.Map{"invoices": invoices})
}

func (h *BillingHandler) Webhook(c *fiber.Ctx) error {
	payload := c.Body()
	provider := c.Params("provider")

	if !h.verifyWebhook(provider, payload, c) {

		if h.alerts != nil {
			_ = h.alerts.AlertFailedWebhook(c.UserContext(), provider, c.Get("X-Event-ID"), "invalid billing webhook signature")
		}

		return writeError(c, fiber.StatusUnauthorized, fmt.Errorf("invalid billing webhook signature"))
	}

	eventID := c.Get("X-Event-ID")
	transition, payloadEventID, eventType, err := parseWebhook(provider, payload)

	if eventID == "" {
		eventID = payloadEventID
	}

	if eventID == "" {

		if h.alerts != nil {
			_ = h.alerts.AlertFailedWebhook(c.UserContext(), provider, "", "missing X-Event-ID header")
		}

		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("X-Event-ID is required"))
	}

	digest := sha256.Sum256(payload)

	if eventType == "" {
		eventType = c.Get("X-Event-Type")
	}

	event := &models.BillingEvent{Provider: models.BillingProvider(provider), ProviderEventID: eventID, EventType: eventType, PayloadHash: hex.EncodeToString(digest[:])}
	var transitionPointer *services.BillingTransition

	if transition.ProviderSubscription != "" && transition.Status != "" {
		transitionPointer = &transition
	}

	created, err := h.billing.ProcessWebhook(c.UserContext(), event, transitionPointer)

	if err != nil {

		if h.alerts != nil {
			_ = h.alerts.AlertFailedWebhook(c.UserContext(), provider, eventID, err.Error())
		}

		return writeError(c, fiber.StatusBadRequest, err)
	}

	_ = created
	return c.SendStatus(fiber.StatusAccepted)
}

func (h *BillingHandler) verifyWebhook(provider string, payload []byte, c *fiber.Ctx) bool {
	signature := c.Get("X-Signature")

	switch strings.ToLower(provider) {
	case string(models.BillingProviderPolar):
		secret := h.polarSecret

		if secret == "" {
			secret = h.webhookSecret
		}

		return secret != "" && billing.VerifyHMACSHA256(payload, signature, secret)
	case string(models.BillingProviderPaystack):
		return h.paystack != nil && h.paystack.VerifyWebhook(payload, c.Get("X-Paystack-Signature"))
	default:
		return false
	}
}

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

	status := subscriptionStatus(eventType, stringValue(data, "status"))
	periodEnd := timeValue(data, "current_period_end", "period_end")
	attemptKeys := []string{"attempts_remaining", "attemptsRemaining", "retries_remaining"}
	transition := services.BillingTransition{Provider: models.BillingProvider(strings.ToLower(provider)), ProviderSubscription: stringValue(data, "subscription_id", "subscription_code", "id"), ProviderCustomer: stringValue(data, "customer_id", "customer_code"), ProviderProduct: stringValue(data, "product_id"), ProviderInvoice: stringValue(data, "invoice_id", "order_id", "invoice_code", "reference"), InvoiceURL: stringValue(data, "invoice_url", "receipt_url", "order_url"), Status: status, CurrentPeriodEnd: periodEnd, CancelAtPeriodEnd: boolValue(data, "cancel_at_period_end", "cancelled"), EventType: eventType, AmountMinor: int64Value(data, "amount", "amount_minor", "amountMinor"), Currency: stringValue(data, "currency"), AttemptsRemaining: intValue(data, attemptKeys...), AttemptsKnown: hasIntValue(data, attemptKeys...), PreviousPlan: stringValue(data, "previous_plan", "previousPlan"), PaidAt: timeValue(data, "paid_at", "paidAt", "created_at")}

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

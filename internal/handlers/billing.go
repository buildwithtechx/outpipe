package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

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

func (h *BillingHandler) Plans(c *fiber.Ctx) error {
	plans, err := h.billing.ListPlans(c.UserContext())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{"plans": plans})
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

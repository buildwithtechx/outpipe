package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type WebhookHandler struct{ webhooks *services.WebhookService }

type CreateWebhookRequest struct {
	Name   string   `json:"name" validate:"required,max=120"`
	URL    string   `json:"url" validate:"required,max=2048"`
	Events []string `json:"events"`
}

func NewWebhookHandler(webhooks *services.WebhookService) (*WebhookHandler, error) {

	if webhooks == nil {
		return nil, fmt.Errorf("webhook service is required")
	}

	return &WebhookHandler{webhooks: webhooks}, nil
}

func (h *WebhookHandler) Create(c *fiber.Ctx) error {
	var input CreateWebhookRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode webhook request: %w", err))
	}

	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	subscription, secret, err := h.webhooks.Create(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), input.Name, input.URL, input.Events)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"subscription": subscription, "secret": secret})
}

func (h *WebhookHandler) List(c *fiber.Ctx) error {
	subscriptions, err := h.webhooks.List(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(subscriptions)
}

func (h *WebhookHandler) Delete(c *fiber.Ctx) error {

	if err := h.webhooks.Delete(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("webhookID"))); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WebhookHandler) Deliveries(c *fiber.Ctx) error {
	deliveries, err := h.webhooks.Deliveries(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("webhookID")))

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(deliveries)
}

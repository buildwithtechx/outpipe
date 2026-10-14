package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

type SubscriberTokenInput struct {
	Token string `json:"token"`
}

func (h *UptimeHandler) ConfirmSubscription(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	var input SubscriberTokenInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if strings.TrimSpace(input.Token) == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("confirmation token is required"))
	}
	if _, err := h.uptime.ConfirmSubscription(c.UserContext(), slug, input.Token); err != nil {
		if errors.Is(err, services.ErrConfirmTokenInvalid) {
			return writeError(c, fiber.StatusGone, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(fiber.Map{"confirmed": true})
}

func (h *UptimeHandler) UnsubscribeWithToken(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	var input SubscriberTokenInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if strings.TrimSpace(input.Token) == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("unsubscribe token is required"))
	}
	if _, err := h.uptime.UnsubscribeWithToken(c.UserContext(), slug, input.Token); err != nil {
		if errors.Is(err, services.ErrUnsubscribeInvalid) {
			return writeError(c, fiber.StatusGone, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(fiber.Map{"unsubscribed": true})
}

func (h *UptimeHandler) ConfirmSubscriptionByPage(c *fiber.Ctx) error {
	page, err := h.uptime.GetStatusPageByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	var input SubscriberTokenInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if strings.TrimSpace(input.Token) == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("confirmation token is required"))
	}
	if _, err := h.uptime.ConfirmSubscription(c.UserContext(), page.Slug, input.Token); err != nil {
		if errors.Is(err, services.ErrConfirmTokenInvalid) {
			return writeError(c, fiber.StatusGone, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(fiber.Map{"confirmed": true})
}

func (h *UptimeHandler) UnsubscribeByPage(c *fiber.Ctx) error {
	page, err := h.uptime.GetStatusPageByID(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	var input SubscriberTokenInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if strings.TrimSpace(input.Token) == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("unsubscribe token is required"))
	}
	if _, err := h.uptime.UnsubscribeWithToken(c.UserContext(), page.Slug, input.Token); err != nil {
		if errors.Is(err, services.ErrUnsubscribeInvalid) {
			return writeError(c, fiber.StatusGone, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(fiber.Map{"unsubscribed": true})
}

func (h *UptimeHandler) ListSubscribers(c *fiber.Ctx) error {
	page, ok := c.Locals("statusPage").(models.UptimeStatusPage)
	if !ok {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("status page not found"))
	}
	limit := 25
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("limit must be between 1 and 100"))
		}
		limit = parsed
	}
	result, err := h.uptime.ListManagedSubscribers(c.UserContext(), page.ID, c.Query("q"), c.Query("status"), limit, c.Query("cursor"))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(result)
}

func (h *UptimeHandler) RemoveSubscriber(c *fiber.Ctx) error {
	page, ok := c.Locals("statusPage").(models.UptimeStatusPage)
	if !ok {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("status page not found"))
	}
	subscriberID := strings.TrimSpace(c.Params("subscriberID"))
	if subscriberID == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("subscriber id is required"))
	}
	if err := h.uptime.RemoveSubscriber(c.UserContext(), page.ID, subscriberID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return writeError(c, fiber.StatusNotFound, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UptimeHandler) ManagerUnsubscribe(c *fiber.Ctx) error {
	page, ok := c.Locals("statusPage").(models.UptimeStatusPage)
	if !ok {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("status page not found"))
	}
	subscriberID := strings.TrimSpace(c.Params("subscriberID"))
	if subscriberID == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("subscriber id is required"))
	}
	view, err := h.uptime.ManagerUnsubscribe(c.UserContext(), page.ID, subscriberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return writeError(c, fiber.StatusNotFound, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(view)
}

func (h *UptimeHandler) UpdateSubscriberSettings(c *fiber.Ctx) error {
	page, ok := c.Locals("statusPage").(models.UptimeStatusPage)
	if !ok {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("status page not found"))
	}
	var input map[string]any
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if len(input) != 1 {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("body must contain exactly subscriptionsEnabled"))
	}
	raw, present := input["subscriptionsEnabled"]
	if !present {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("body must contain exactly subscriptionsEnabled"))
	}
	enabled, isBool := raw.(bool)
	if !isBool {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("subscriptionsEnabled must be a boolean"))
	}
	updated, err := h.uptime.UpdateSubscriberSettings(c.UserContext(), page.ID, enabled)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return writeError(c, fiber.StatusNotFound, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(fiber.Map{"subscriptionsEnabled": updated.SubscriptionsEnabled})
}

type SubscriberScopeInput struct {
	ComponentIDs []string `json:"componentIds"`
}

func (h *UptimeHandler) SetSubscriberScope(c *fiber.Ctx) error {
	page, ok := c.Locals("statusPage").(models.UptimeStatusPage)
	if !ok {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("status page not found"))
	}
	subscriberID := strings.TrimSpace(c.Params("subscriberID"))
	if subscriberID == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("subscriber id is required"))
	}
	var input SubscriberScopeInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	view, err := h.uptime.SetSubscriberScope(c.UserContext(), page.ID, subscriberID, input.ComponentIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, repositories.ErrNotFound) {
			return writeError(c, fiber.StatusNotFound, err)
		}
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(view)
}

package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

type UsageHandler struct{ usage *services.UsageService }

func NewUsageHandler(usage *services.UsageService) (*UsageHandler, error) {

	if usage == nil {
		return nil, fmt.Errorf("usage service is required")
	}

	return &UsageHandler{usage: usage}, nil
}

func (h *UsageHandler) Ingest(c *fiber.Ctx) error {
	var event models.UsageEvent

	if err := c.BodyParser(&event); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	if err := h.usage.Record(c.UserContext(), &event); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func (h *UsageHandler) Events(c *fiber.Ctx) error {
	from, to, err := usagePeriod(c)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	events, err := h.usage.ListEvents(c.UserContext(), c.Params("organizationID"), from, to)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(fiber.Map{"from": from, "to": to, "events": events})
}

func (h *UsageHandler) Snapshot(c *fiber.Ctx) error {
	periodStart := time.Now().UTC().Truncate(time.Hour).Add(-time.Hour)

	if value := c.Query("periodStart"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)

		if err != nil {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("invalid periodStart: %w", err))
		}

		periodStart = parsed
	}

	snapshot, err := h.usage.FindSnapshot(c.UserContext(), c.Params("organizationID"), periodStart)

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(snapshot)
}

func usagePeriod(c *fiber.Ctx) (time.Time, time.Time, error) {
	to := time.Now().UTC()
	from := to.Add(-24 * time.Hour)
	var err error

	if value := c.Query("from"); value != "" {
		from, err = time.Parse(time.RFC3339, value)

		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid from: %w", err)
		}
	}

	if value := c.Query("to"); value != "" {
		to, err = time.Parse(time.RFC3339, value)

		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid to: %w", err)
		}
	}

	if !to.After(from) || to.Sub(from) > 31*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("usage period must be positive and no longer than 31 days")
	}

	return from, to, nil
}

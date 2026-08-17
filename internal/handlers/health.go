package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	ready func(context.Context) error
}

func NewHealthHandler(ready func(context.Context) error) *HealthHandler {
	return &HealthHandler{ready: ready}
}

func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *HealthHandler) Readiness(c *fiber.Ctx) error {

	if h.ready == nil {
		return c.JSON(fiber.Map{"status": "ok"})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	if err := h.ready(ctx); err != nil {
		return writeError(c, fiber.StatusServiceUnavailable, err)
	}

	return c.JSON(fiber.Map{"status": "ready"})
}

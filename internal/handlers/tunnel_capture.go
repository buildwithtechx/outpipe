package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/repositories"
)

func (h *TunnelHandler) SetCapture(c *fiber.Ctx) error {
	var input struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.BodyParser(&input); err != nil || input.Enabled == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	tunnel, err := h.tunnels.SetCapture(c.UserContext(), c.Params("tunnelID"), *input.Enabled)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(tunnel)
}

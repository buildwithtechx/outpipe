package handlers

import "github.com/gofiber/fiber/v2"

func (h *TunnelHandler) SetCapture(c *fiber.Ctx) error {
	var input struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.BodyParser(&input); err != nil || input.Enabled == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	tunnel, err := h.tunnels.SetCapture(c.UserContext(), c.Params("tunnelID"), *input.Enabled)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(tunnel)
}

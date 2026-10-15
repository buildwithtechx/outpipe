package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/repositories"
)

func (h *TunnelHandler) SetCapture(c *fiber.Ctx) error {
	var input struct {
		Enabled *bool `json:"enabled"`
		Policy  *struct {
			Rules string `json:"rules"`
		} `json:"policy"`
	}
	if err := c.BodyParser(&input); err != nil || (input.Enabled == nil && input.Policy == nil) {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if input.Policy != nil {
		if _, err := h.tunnels.SetCapturePolicy(c.UserContext(), c.Params("tunnelID"), input.Policy.Rules); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return c.SendStatus(fiber.StatusNotFound)
			}
			return writeError(c, fiber.StatusBadRequest, err)
		}
	}
	if input.Enabled != nil {
		tunnel, err := h.tunnels.SetCapture(c.UserContext(), c.Params("tunnelID"), *input.Enabled)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return c.SendStatus(fiber.StatusNotFound)
			}
			return writeError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(tunnel)
	}
	policy, err := h.tunnels.GetCapturePolicy(c.UserContext(), c.Params("tunnelID"))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(policy)
}

func (h *TunnelHandler) GetCapturePolicy(c *fiber.Ctx) error {
	policy, err := h.tunnels.GetCapturePolicy(c.UserContext(), c.Params("tunnelID"))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(policy)
}

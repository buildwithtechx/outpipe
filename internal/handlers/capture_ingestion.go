package handlers

import (
	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

func (h *ObservabilityHandler) SetTunnels(tunnels *services.TunnelService) { h.tunnels = tunnels }

func (h *ObservabilityHandler) IngestCapture(c *fiber.Ctx) error {
	if h.tunnels == nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}
	if len(c.Body()) > 512*1024 {
		return c.SendStatus(fiber.StatusRequestEntityTooLarge)
	}
	var capture models.RequestCapture
	if err := c.BodyParser(&capture); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	tunnel, err := h.tunnels.Find(c.UserContext(), capture.TunnelID)
	if err != nil || tunnel.OrganizationID != capture.OrganizationID || !tunnel.CaptureEnabled || tunnel.Status == models.TunnelStatusRevoked {
		return c.SendStatus(fiber.StatusForbidden)
	}
	capture.ID = ""
	if err := h.svc.IngestCapture(c.UserContext(), &capture); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusCreated)
}

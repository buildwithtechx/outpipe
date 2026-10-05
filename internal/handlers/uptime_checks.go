package handlers

import "github.com/gofiber/fiber/v2"

func (h *UptimeHandler) MonitorChecks(c *fiber.Ctx) error {
	rows, err := h.uptime.MonitorChecks(c.UserContext(), c.Params("organizationID"), c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	return c.JSON(rows)
}

package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/repositories"
)

func (h *UptimeHandler) MonitorChecks(c *fiber.Ctx) error {
	rows, err := h.uptime.MonitorChecks(c.UserContext(), c.Params("organizationID"), c.Params("id"))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(rows)
}

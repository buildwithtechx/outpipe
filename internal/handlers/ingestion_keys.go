package handlers

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

func (h *APIKeyHandler) CreateIngestionKey(c *fiber.Ctx) error {
	userID, err := sessionUserID(c)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	expires := time.Now().UTC().Add(time.Hour)
	raw, key, err := h.keys.CreateForOrganization(c.UserContext(), userID, c.Params("organizationID"), "Telemetry ingestion", []string{"telemetry:write"}, &expires, "otlp")
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"key": key, "raw": raw})
}

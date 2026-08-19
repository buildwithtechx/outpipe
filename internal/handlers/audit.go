package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type AuditLogHandler struct{ audit *services.AuditService }

func NewAuditLogHandler(audit *services.AuditService) (*AuditLogHandler, error) {

	if audit == nil {
		return nil, fmt.Errorf("audit service is required")
	}

	return &AuditLogHandler{audit: audit}, nil
}

func (h *AuditLogHandler) List(c *fiber.Ctx) error {
	to := time.Now().UTC()
	from := to.Add(-30 * 24 * time.Hour)
	limit := 100

	if err := parseTimeQuery(c, "from", &from); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	if err := parseTimeQuery(c, "to", &to); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 || parsed > 1000 {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("limit must be between 1 and 1000"))
		}

		limit = parsed
	}

	if !to.After(from) || to.Sub(from) > 90*24*time.Hour {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("audit period must be positive and no longer than 90 days"))
	}

	events, err := h.audit.ListByOrganization(c.UserContext(), c.Params("organizationID"), from, to, limit)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(fiber.Map{"from": from, "to": to, "events": events})
}

func parseTimeQuery(c *fiber.Ctx, name string, target *time.Time) error {
	value := c.Query(name)

	if value == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, value)

	if err != nil {
		return fmt.Errorf("invalid %s: %w", name, err)
	}

	*target = parsed
	return nil
}

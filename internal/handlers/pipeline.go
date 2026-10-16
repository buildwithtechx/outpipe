package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type PipelineHandler struct {
	health *services.PipelineHealthService
}

func NewPipelineHandler(health *services.PipelineHealthService) (*PipelineHandler, error) {
	if health == nil {
		return nil, fmt.Errorf("pipeline health service is required")
	}
	return &PipelineHandler{health: health}, nil
}

func (h *PipelineHandler) Health(c *fiber.Ctx) error {
	health, err := h.health.Health(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(health)
}

func (h *PipelineHandler) DeadLetters(c *fiber.Ctx) error {
	limit := 25
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("limit must be between 1 and 100"))
		}
		limit = parsed
	}
	letters, err := h.health.DeadLetters(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), limit)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(letters)
}

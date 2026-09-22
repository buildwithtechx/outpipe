package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type SupportHandler struct{ support *services.SupportService }

type ContactRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Topic   string `json:"topic"`
	Message string `json:"message" validate:"required"`
}

type BugReportRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Category     string `json:"category"`
	Summary      string `json:"summary" validate:"required"`
	Reproduction string `json:"reproduction"`
	Expected     string `json:"expected"`
	Actual       string `json:"actual" validate:"required"`
}

func NewSupportHandler(support *services.SupportService) (*SupportHandler, error) {
	if support == nil {
		return nil, fmt.Errorf("support service is required")
	}
	return &SupportHandler{support: support}, nil
}

func (h *SupportHandler) Contact(c *fiber.Ctx) error {
	var input ContactRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode contact request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	if err := h.support.Contact(c.UserContext(), input.Name, input.Email, input.Topic, input.Message); err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.SendStatus(fiber.StatusAccepted)
}

func (h *SupportHandler) BugReport(c *fiber.Ctx) error {
	var input BugReportRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode bug report request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	if err := h.support.BugReport(c.UserContext(), input.Name, input.Email, input.Category, input.Summary, input.Reproduction, input.Expected, input.Actual); err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.SendStatus(fiber.StatusAccepted)
}

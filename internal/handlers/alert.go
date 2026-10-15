package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type AlertHandler struct{ bridge *services.AlertBridgeService }

func NewAlertHandler(bridge *services.AlertBridgeService) (*AlertHandler, error) {
	if bridge == nil {
		return nil, fmt.Errorf("alert bridge service is required")
	}
	return &AlertHandler{bridge: bridge}, nil
}

type CreateAlertRequest struct {
	Name                 string   `json:"name"`
	MonitorIDs           []string `json:"monitorIds"`
	ConsecutiveFailures  int      `json:"consecutiveFailures"`
	ConsecutiveRecovered int      `json:"consecutiveRecoveries"`
	EvalIntervalSeconds  int      `json:"evalIntervalSeconds"`
	StatusPageTitle      string   `json:"statusPageTitle"`
	StatusPageImpact     string   `json:"statusPageImpact"`
}

func (h *AlertHandler) Create(c *fiber.Ctx) error {
	var input CreateAlertRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode alert request: %w", err))
	}
	alert, err := h.bridge.Create(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), services.CreateAlertInput{
		Name:                 input.Name,
		MonitorIDs:           input.MonitorIDs,
		ConsecutiveFailures:  input.ConsecutiveFailures,
		ConsecutiveRecovered: input.ConsecutiveRecovered,
		EvalIntervalSeconds:  input.EvalIntervalSeconds,
		StatusPageTitle:      input.StatusPageTitle,
		StatusPageImpact:     input.StatusPageImpact,
	})
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(alert)
}

func (h *AlertHandler) List(c *fiber.Ctx) error {
	alerts, err := h.bridge.List(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(alerts)
}

func (h *AlertHandler) Delete(c *fiber.Ctx) error {
	if err := h.bridge.Delete(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), c.Params("alertID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type SetAlertStatusPageRequest struct {
	PublicTitle  string   `json:"publicTitle"`
	Impact       string   `json:"impact"`
	ComponentIDs []string `json:"componentIds"`
}

func (h *AlertHandler) SetStatusPage(c *fiber.Ctx) error {
	var input SetAlertStatusPageRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode alert status page request: %w", err))
	}
	alert, err := h.bridge.SetStatusPageLink(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), c.Params("alertID"), input.PublicTitle, input.Impact, input.ComponentIDs)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(alert)
}

type SetAlertEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *AlertHandler) SetEnabled(c *fiber.Ctx) error {
	var input SetAlertEnabledRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode alert request: %w", err))
	}
	alert, err := h.bridge.SetEnabled(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), c.Params("alertID"), input.Enabled)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.JSON(alert)
}

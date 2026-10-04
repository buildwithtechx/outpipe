package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

type UptimeHandler struct {
	uptime *services.UptimeService
}

func NewUptimeHandler(uptime *services.UptimeService) (*UptimeHandler, error) {
	if uptime == nil {
		return nil, fmt.Errorf("uptime service is required")
	}
	return &UptimeHandler{uptime: uptime}, nil
}

func (h *UptimeHandler) ListMonitors(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	monitors, err := h.uptime.ListOrgMonitors(c.UserContext(), orgID)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(monitors)
}

func (h *UptimeHandler) CreateMonitor(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	var input services.CreateMonitorInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	monitor, err := h.uptime.CreateMonitor(c.UserContext(), orgID, input)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(monitor)
}

func (h *UptimeHandler) GetMonitor(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	monitor, err := h.uptime.GetMonitor(c.UserContext(), id)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.JSON(monitor)
}

func (h *UptimeHandler) DeleteMonitor(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if err := h.uptime.DeleteMonitor(c.UserContext(), id); err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *UptimeHandler) ProbeMonitor(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	monitor, err := h.uptime.GetMonitor(c.UserContext(), id)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	check, err := h.uptime.Probe(c.UserContext(), monitor)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(check)
}

func (h *UptimeHandler) ListIncidents(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	incidents, err := h.uptime.ListOrgIncidents(c.UserContext(), orgID)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(incidents)
}

func (h *UptimeHandler) CreateIncident(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	var input services.CreateIncidentInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	incident, err := h.uptime.CreateIncident(c.UserContext(), orgID, input)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(incident)
}

type AddUpdateInput struct {
	Status  models.IncidentStatus `json:"status"`
	Message string                `json:"message"`
}

func (h *UptimeHandler) AddIncidentUpdate(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	var input AddUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if err := h.uptime.AddIncidentUpdate(c.UserContext(), id, input.Status, input.Message); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

type UpsertStatusPageInput struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CustomDomain string `json:"customDomain"`
}

func (h *UptimeHandler) GetStatusPage(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	page, err := h.uptime.GetStatusPageByOrg(c.UserContext(), orgID)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.JSON(page)
}

func (h *UptimeHandler) UpsertStatusPage(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	var input UpsertStatusPageInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	page, err := h.uptime.UpsertStatusPage(c.UserContext(), orgID, input.Slug, input.Title, input.Description, input.CustomDomain)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(page)
}

func (h *UptimeHandler) GetPublicStatus(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	data, err := h.uptime.GetPublicStatusData(c.UserContext(), slug)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.JSON(data)
}

type SubscribeInput struct {
	Email string `json:"email"`
}

func (h *UptimeHandler) Subscribe(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	var input SubscribeInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode request: %w", err))
	}
	if err := h.uptime.Subscribe(c.UserContext(), slug, input.Email); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"subscribed": true})
}

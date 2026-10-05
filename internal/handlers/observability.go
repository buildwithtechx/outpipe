package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

type ObservabilityHandler struct {
	svc     *services.ObservabilityService
	tunnels *services.TunnelService
}

func NewObservabilityHandler(svc *services.ObservabilityService) *ObservabilityHandler {
	return &ObservabilityHandler{svc: svc}
}

func (h *ObservabilityHandler) GetStats(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	timeRange := c.Query("range", "24h")
	stats, err := h.svc.GetStats(c.Context(), orgID, timeRange)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("fetch stats: %v", err),
		})
	}
	return c.JSON(stats)
}

func (h *ObservabilityHandler) ListTraces(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	traces, err := h.svc.ListTraces(c.Context(), orgID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("list traces: %v", err),
		})
	}
	return c.JSON(traces)
}

func (h *ObservabilityHandler) GetTraceWaterfall(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	traceID := c.Params("traceId")
	spans, err := h.svc.GetTraceWaterfall(c.Context(), orgID, traceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("get trace waterfall: %v", err),
		})
	}
	return c.JSON(spans)
}

func (h *ObservabilityHandler) ListLogs(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	search := c.Query("search")
	severity := c.Query("severity")
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	logs, err := h.svc.ListLogs(c.Context(), orgID, search, severity, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("list logs: %v", err),
		})
	}
	return c.JSON(logs)
}

func (h *ObservabilityHandler) ListCaptures(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	tunnelID := c.Query("tunnel_id")
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	captures, err := h.svc.ListCaptures(c.Context(), orgID, tunnelID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("list captures: %v", err),
		})
	}
	return c.JSON(captures)
}

func (h *ObservabilityHandler) GetCapture(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	captureID := c.Params("captureId")
	capture, err := h.svc.GetCapture(c.Context(), orgID, captureID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("capture not found: %v", err),
		})
	}
	return c.JSON(capture)
}

func (h *ObservabilityHandler) ReplayRequest(c *fiber.Ctx) error {
	var input models.ReplayRequestInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("parse replay payload: %v", err),
		})
	}
	result, err := h.svc.ReplayRequest(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("replay request: %v", err),
		})
	}
	return c.Status(http.StatusOK).JSON(result)
}

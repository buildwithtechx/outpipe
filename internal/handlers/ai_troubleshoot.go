package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

type AIAgentHandler struct {
	svc *services.AIAgentService
}

func NewAIAgentHandler(svc *services.AIAgentService) (*AIAgentHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("ai agent service is required")
	}
	return &AIAgentHandler{svc: svc}, nil
}

func (h *AIAgentHandler) ListThreads(c *fiber.Ctx) error {
	orgID := c.Params("organizationID")
	userID, _ := c.Locals("userID").(string)

	threads, err := h.svc.ListThreads(c.Context(), orgID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("list ai threads: %v", err),
		})
	}
	cfg := h.svc.GetConfig()
	return c.JSON(fiber.Map{
		"threads":            threads,
		"configured":         cfg.Enabled,
		"provider":           cfg.Provider,
		"model":              cfg.Model,
		"availableProviders": cfg.AvailableProviders,
	})
}

func (h *AIAgentHandler) DeleteThread(c *fiber.Ctx) error {
	orgID := c.Params("organizationID")
	threadID := c.Params("threadID")

	if err := h.svc.DeleteThread(c.Context(), orgID, threadID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("delete ai thread: %v", err),
		})
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *AIAgentHandler) ChatStream(c *fiber.Ctx) error {
	orgID := c.Params("organizationID")
	userID, _ := c.Locals("userID").(string)

	var req models.AIChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message is required",
		})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		emit := func(event models.AIStreamEvent) error {
			data, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
				return err
			}
			return w.Flush()
		}
		_ = h.svc.InvestigateStream(ctx, orgID, userID, req, emit)
	})

	return nil
}

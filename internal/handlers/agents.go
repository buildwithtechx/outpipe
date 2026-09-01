package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

type AgentHandler struct{ agents *services.AgentService }

func (h *AgentHandler) OrganizationID(ctx context.Context, id string) (string, error) {
	agent, err := h.agents.Find(ctx, id)
	return agent.OrganizationID, err
}

type RegisterAgentRequest struct {
	Name string `json:"name"`
}

type AgentHeartbeatRequest struct {
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
	Platform string `json:"platform"`
}

func NewAgentHandler(agents *services.AgentService) (*AgentHandler, error) {

	if agents == nil {
		return nil, fmt.Errorf("agent service is required")
	}

	return &AgentHandler{agents: agents}, nil
}

func (h *AgentHandler) Register(c *fiber.Ctx) error {
	var input RegisterAgentRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode agent request: %w", err))
	}

	token, agent, err := h.agents.Register(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), input.Name)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"agent": agent, "token": token})
}

func (h *AgentHandler) List(c *fiber.Ctx) error {
	agents, err := h.agents.List(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(agents)
}

func (h *AgentHandler) Heartbeat(c *fiber.Ctx) error {
	var input AgentHeartbeatRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode heartbeat request: %w", err))
	}

	if err := h.agents.Heartbeat(c.UserContext(), c.Params("agentID"), input.Version, input.Hostname, input.Platform); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AgentHandler) Revoke(c *fiber.Ctx) error {

	if err := h.agents.Revoke(c.UserContext(), c.Params("agentID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AgentHandler) Authenticate(c *fiber.Ctx) error {
	value := strings.TrimSpace(c.Get("Authorization"))
	parts := strings.SplitN(value, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return writeError(c, fiber.StatusUnauthorized, fmt.Errorf("bearer agent token is required"))
	}

	agent, plan, err := h.agents.AuthenticateWithPlan(c.UserContext(), strings.TrimSpace(parts[1]))

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	return c.JSON(fiber.Map{"agentId": agent.ID, "organizationId": agent.OrganizationID, "status": models.AgentStatusOnline, "limits": fiber.Map{"maxTunnels": plan.MaxTunnels, "maxConnections": plan.MaxConnections, "bandwidthBytes": plan.BandwidthBytes}})
}

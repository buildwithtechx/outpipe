package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/security"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type TunnelHandler struct{ tunnels *services.TunnelService }

func (h *TunnelHandler) OrganizationID(ctx context.Context, id string) (string, error) {
	tunnel, err := h.tunnels.Find(ctx, id)
	return tunnel.OrganizationID, err
}

type CreateTunnelRequest struct {
	Name           string                `json:"name" validate:"required,max=120"`
	Protocol       models.TunnelProtocol `json:"protocol" validate:"required"`
	TargetHost     string                `json:"targetHost" validate:"required,max=253"`
	TargetPort     int                   `json:"targetPort" validate:"gte=1,lte=65535"`
	PublicHostname string                `json:"publicHostname,omitempty" validate:"max=253"`
	Password       string                `json:"password,omitempty" validate:"omitempty,min=8,max=256"`
}

type UpdateTunnelStatusRequest struct {
	Status models.TunnelStatus `json:"status"`
}

func NewTunnelHandler(tunnels *services.TunnelService) (*TunnelHandler, error) {

	if tunnels == nil {
		return nil, fmt.Errorf("tunnel service is required")
	}

	return &TunnelHandler{tunnels: tunnels}, nil
}

func (h *TunnelHandler) Create(c *fiber.Ctx) error {
	var input CreateTunnelRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode tunnel request: %w", err))
	}

	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	tunnel, err := h.tunnels.Create(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), input.Name, input.Protocol, input.TargetHost, input.TargetPort, input.PublicHostname, input.Password)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(tunnel)
}

func (h *TunnelHandler) List(c *fiber.Ctx) error {
	tunnels, err := h.tunnels.List(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(tunnels)
}

func (h *TunnelHandler) Inspect(c *fiber.Ctx) error {
	tunnel, err := h.tunnels.Find(c.UserContext(), strings.TrimSpace(c.Params("tunnelID")))

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(tunnel)
}

func (h *TunnelHandler) Policy(c *fiber.Ctx) error {
	tunnel, err := h.tunnels.Policy(c.UserContext(), strings.TrimSpace(c.Params("tunnelID")))

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(fiber.Map{"organizationId": tunnel.OrganizationID, "publicHostname": tunnel.PublicHostname, "status": tunnel.Status, "passwordProtected": tunnel.PasswordHash != ""})
}

func (h *TunnelHandler) VerifyPassword(c *fiber.Ctx) error {
	var input struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode tunnel password request: %w", err))
	}

	if strings.TrimSpace(input.Password) == "" || len(input.Password) < 8 || len(input.Password) > 256 {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("tunnel password must be between 8 and 256 bytes"))
	}

	tunnel, err := h.tunnels.Policy(c.UserContext(), strings.TrimSpace(c.Params("tunnelID")))

	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.JSON(fiber.Map{"valid": tunnel.PasswordHash != "" && security.VerifyPassword(input.Password, tunnel.PasswordHash)})
}

func (h *TunnelHandler) SetStatus(c *fiber.Ctx) error {
	var input UpdateTunnelStatusRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode tunnel status request: %w", err))
	}

	if err := h.tunnels.SetStatus(c.UserContext(), c.Params("tunnelID"), input.Status); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TunnelHandler) Revoke(c *fiber.Ctx) error {

	if err := h.tunnels.Revoke(c.UserContext(), c.Params("tunnelID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

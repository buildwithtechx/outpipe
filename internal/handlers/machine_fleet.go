package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type MachineFleetHandler struct {
	fleets *services.MachineFleetService
}

func NewMachineFleetHandler(fleets *services.MachineFleetService) (*MachineFleetHandler, error) {
	if fleets == nil {
		return nil, fmt.Errorf("machine fleet service is required")
	}
	return &MachineFleetHandler{fleets: fleets}, nil
}

type CreateFleetRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *MachineFleetHandler) Create(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	var input CreateFleetRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode fleet request: %w", err))
	}
	fleet, err := h.fleets.CreateFleet(c.UserContext(), orgID, input.Name, input.Description)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fleet)
}

func (h *MachineFleetHandler) List(c *fiber.Ctx) error {
	fleets, err := h.fleets.ListFleets(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(fleets)
}

func (h *MachineFleetHandler) Status(c *fiber.Ctx) error {
	status, err := h.fleets.Status(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID")))
	if err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.JSON(status)
}

func (h *MachineFleetHandler) Delete(c *fiber.Ctx) error {
	if err := h.fleets.DeleteFleet(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID"))); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type AttachFleetTokenRequest struct {
	MachineTokenID string `json:"machineTokenId"`
	Label          string `json:"label"`
}

func (h *MachineFleetHandler) AttachToken(c *fiber.Ctx) error {
	var input AttachFleetTokenRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode attach request: %w", err))
	}
	attachment, err := h.fleets.AttachToken(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID")), strings.TrimSpace(input.MachineTokenID), input.Label)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(attachment)
}

func (h *MachineFleetHandler) BulkCreateTunnels(c *fiber.Ctx) error {
	var specs []services.FleetTunnelSpec
	if err := c.BodyParser(&specs); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode fleet tunnel specs: %w", err))
	}
	userID, err := sessionUserID(c)
	if err != nil {
		userID = ""
	}
	tunnels, err := h.fleets.BulkCreateTunnels(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID")), userID, specs)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.Status(fiber.StatusCreated).JSON(tunnels)
}

func (h *MachineFleetHandler) RotateTokens(c *fiber.Ctx) error {
	userID, err := sessionUserID(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	rotated, err := h.fleets.RotateFleetTokens(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID")), userID)
	type rotatedToken struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Prefix string `json:"prefix"`
		Raw    string `json:"raw"`
	}
	out := make([]rotatedToken, 0, len(rotated))
	for _, item := range rotated {
		out = append(out, rotatedToken{ID: item.Token.ID, Name: item.Token.Name, Prefix: item.Token.Prefix, Raw: item.Raw})
	}
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"rotated": out,
		})
	}
	return c.JSON(fiber.Map{"rotated": out})
}

func (h *MachineFleetHandler) Revoke(c *fiber.Ctx) error {
	if err := h.fleets.RevokeFleet(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID"))); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MachineFleetHandler) Reconcile(c *fiber.Ctx) error {
	result, err := h.fleets.Reconcile(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), strings.TrimSpace(c.Params("fleetID")))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(result)
}

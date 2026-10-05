package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

func (h *TunnelHandler) SetMachineService(secrets *services.SecretService) { h.secrets = secrets }
func (h *TunnelHandler) SetMachineSigningKey(key string)                   { h.signingKey = key }

func (h *TunnelHandler) CreateMachine(c *fiber.Ctx) error {
	if h.secrets == nil || h.signingKey == "" {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}
	parts := strings.Fields(c.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	machine, err := h.secrets.VerifyMachineToken(c.UserContext(), parts[1])
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	var scopes []string
	if err := json.Unmarshal([]byte(machine.Scopes), &scopes); err != nil {
		return c.SendStatus(fiber.StatusForbidden)
	}
	allowed := false
	for _, scope := range scopes {
		if scope == "tunnels:write" {
			allowed = true
		}
	}
	if !allowed || machine.ProjectID != nil || machine.EnvironmentID != nil {
		return c.SendStatus(fiber.StatusForbidden)
	}
	var input CreateTunnelRequest
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	if input.TargetHost != "127.0.0.1" && input.TargetHost != "localhost" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "machine tunnels must target the local machine"})
	}
	tunnel, err := h.tunnels.CreateWithMachineOwner(c.UserContext(), machine.OrganizationID, input.Name, input.Protocol, input.TargetHost, input.TargetPort, input.PublicHostname, input.Password, string(input.Metadata), &machine.ID)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	expires := time.Now().UTC().Add(15 * time.Minute)
	if machine.ExpiresAt.Before(expires) {
		expires = *machine.ExpiresAt
	}
	token, err := auth.SignRelayToken(auth.RelayClaims{Sub: machine.ID, Org: machine.OrganizationID, TunnelID: tunnel.ID, MachineTokenID: machine.ID, MaxTunnels: 1, Exp: expires.Unix(), Iat: time.Now().Unix()}, h.signingKey)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"tunnel": tunnel, "relayToken": token, "expiresAt": expires})
}

func (h *TunnelHandler) machinePolicyAllowed(c *fiber.Ctx, tunnel models.Tunnel) bool {
	if tunnel.MachineTokenID == nil {
		return true
	}
	if h.secrets == nil {
		return false
	}
	tokens, err := h.secrets.ListMachineTokens(c.UserContext(), tunnel.OrganizationID)
	if err != nil {
		return false
	}
	for _, token := range tokens {
		if token.ID == *tunnel.MachineTokenID {
			return token.RevokedAt == nil && token.ExpiresAt != nil && token.ExpiresAt.After(time.Now())
		}
	}
	return false
}

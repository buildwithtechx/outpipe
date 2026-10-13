package handlers

import (
	"encoding/json"
	"fmt"
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

func (h *TunnelHandler) machineCredential(c *fiber.Ctx) (models.SecretMachineToken, int, error) {
	if h.secrets == nil || h.signingKey == "" {
		return models.SecretMachineToken{}, fiber.StatusServiceUnavailable, fmt.Errorf("machine tunnel service is unavailable")
	}
	parts := strings.Fields(c.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return models.SecretMachineToken{}, fiber.StatusUnauthorized, fmt.Errorf("machine credential is required")
	}
	machine, err := h.secrets.VerifyMachineToken(c.UserContext(), parts[1])
	if err != nil {
		return models.SecretMachineToken{}, fiber.StatusUnauthorized, err
	}
	var scopes []string
	if err := json.Unmarshal([]byte(machine.Scopes), &scopes); err != nil {
		return models.SecretMachineToken{}, fiber.StatusForbidden, err
	}
	allowed := false
	for _, scope := range scopes {
		if scope == "tunnels:write" {
			allowed = true
		}
	}
	if !allowed || machine.ProjectID != nil || machine.EnvironmentID != nil {
		return models.SecretMachineToken{}, fiber.StatusForbidden, fmt.Errorf("machine credential lacks tunnels:write scope")
	}
	return machine, fiber.StatusOK, nil
}

func (h *TunnelHandler) signMachineRelayToken(machine models.SecretMachineToken, tunnelID string) (string, time.Time, error) {
	expires := time.Now().UTC().Add(15 * time.Minute)
	if machine.ExpiresAt.Before(expires) {
		expires = *machine.ExpiresAt
	}
	token, err := auth.SignRelayToken(auth.RelayClaims{Sub: machine.ID, Org: machine.OrganizationID, TunnelID: tunnelID, MachineTokenID: machine.ID, MaxTunnels: 1, Exp: expires.Unix(), Iat: time.Now().Unix()}, h.signingKey)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (h *TunnelHandler) CreateMachine(c *fiber.Ctx) error {
	machine, status, err := h.machineCredential(c)
	if err != nil {
		return c.SendStatus(status)
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
	token, expires, err := h.signMachineRelayToken(machine, tunnel.ID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"tunnel": tunnel, "relayToken": token, "expiresAt": expires})
}

func (h *TunnelHandler) RefreshMachineRelayToken(c *fiber.Ctx) error {
	machine, status, err := h.machineCredential(c)
	if err != nil {
		return c.SendStatus(status)
	}
	tunnel, err := h.tunnels.Find(c.UserContext(), strings.TrimSpace(c.Params("tunnelID")))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if tunnel.OrganizationID != machine.OrganizationID || tunnel.MachineTokenID == nil || *tunnel.MachineTokenID != machine.ID || tunnel.Status == models.TunnelStatusRevoked || tunnel.RevokedAt != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	token, expires, err := h.signMachineRelayToken(machine, tunnel.ID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"relayToken": token, "expiresAt": expires})
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

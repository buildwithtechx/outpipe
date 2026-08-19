package handlers

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type APIKeyHandler struct{ keys *services.APIKeyService }

type CreateAPIKeyRequest struct {
	Name      string   `json:"name" validate:"required,max=120"`
	Scopes    []string `json:"scopes"`
	ExpiresAt *string  `json:"expiresAt"`
}

func NewAPIKeyHandler(keys *services.APIKeyService) (*APIKeyHandler, error) {

	if keys == nil {
		return nil, fmt.Errorf("api key service is required")
	}

	return &APIKeyHandler{keys: keys}, nil
}

func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	organizationID := strings.TrimSpace(c.Params("organizationID"))
	var scope *string

	if organizationID != "" {
		scope = &organizationID
	}

	keys, err := h.keys.List(c.UserContext(), userID, scope)

	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(keys)
}

func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	var input CreateAPIKeyRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode api key request: %w", err))
	}

	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	for _, scope := range input.Scopes {

		if !validAPIKeyScope(scope) {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("invalid api key scope %q", scope))
		}
	}

	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	var expiresAt *time.Time

	if input.ExpiresAt != nil && strings.TrimSpace(*input.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*input.ExpiresAt))

		if err != nil {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("invalid expiresAt: %w", err))
		}

		if !parsed.After(time.Now().UTC()) {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("expiresAt must be in the future"))
		}

		expiresAt = &parsed
	}

	organizationID := strings.TrimSpace(c.Params("organizationID"))

	if organizationID == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("organization is required"))
	}

	raw, key, err := h.keys.CreateForOrganization(c.UserContext(), userID, organizationID, input.Name, input.Scopes, expiresAt)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"key": key, "token": raw})
}

func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	organizationID := strings.TrimSpace(c.Params("organizationID"))
	var scope *string

	if organizationID != "" {
		scope = &organizationID
	}

	if err := h.keys.RevokeOwned(c.UserContext(), userID, scope, c.Params("apiKeyID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func validAPIKeyScope(scope string) bool {
	return slices.Contains([]string{"organization:read", "organization:write", "organization:admin", "organization:owner", "tunnels:read", "tunnels:write", "agents:read", "agents:write", "domains:read", "domains:write", "account:read", "account:write", "billing:read", "*"}, scope)
}

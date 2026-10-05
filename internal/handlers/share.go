package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type ShareHandler struct {
	shares *services.ShareService
}

func NewShareHandler(shares *services.ShareService) (*ShareHandler, error) {
	if shares == nil {
		return nil, fmt.Errorf("share service is required")
	}
	return &ShareHandler{shares: shares}, nil
}

type RevealShareRequest struct {
	Verifier         string  `json:"verifier" validate:"required"`
	PasswordVerifier *string `json:"passwordVerifier,omitempty"`
}

func (h *ShareHandler) Create(c *fiber.Ctx) error {
	return h.create(c, false)
}

func (h *ShareHandler) CreateOrg(c *fiber.Ctx) error {
	return h.create(c, true)
}

func (h *ShareHandler) create(c *fiber.Ctx, organizationBound bool) error {
	var input services.CreateShareInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode share request: %w", err))
	}
	if input.Ciphertext == "" || input.IV == "" || input.KeyVerifier == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("ciphertext, iv, and keyVerifier are required"))
	}
	if organizationBound {
		userID, err := sessionUserID(c)
		if err != nil {
			return writeError(c, fiber.StatusUnauthorized, err)
		}
		orgID := c.Params("organizationID")
		input.OrganizationID, input.CreatedByID = &orgID, &userID
	} else if input.OrganizationID != nil || input.CreatedByID != nil || input.ProjectID != nil || input.EnvironmentID != nil {
		return writeError(c, fiber.StatusForbidden, fmt.Errorf("organization shares require the authenticated organization endpoint"))
	}
	link, err := h.shares.CreateShare(c.UserContext(), input)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":        link.ID,
		"expiresAt": link.ExpiresAt,
		"maxViews":  link.MaxViews,
	})
}

func (h *ShareHandler) GetMeta(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	meta, err := h.shares.GetShareMeta(c.UserContext(), id)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("share not found or expired"))
	}
	return c.JSON(meta)
}

func (h *ShareHandler) Reveal(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	var input RevealShareRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode reveal request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	result, err := h.shares.RevealShare(c.UserContext(), id, input.Verifier, input.PasswordVerifier)
	if err != nil {
		return writeError(c, fiber.StatusNotFound, fmt.Errorf("share not found, expired, or invalid key"))
	}
	return c.JSON(result)
}

func (h *ShareHandler) ListOrgShares(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	shares, err := h.shares.ListOrgShares(c.UserContext(), orgID)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(shares)
}

func (h *ShareHandler) RevokeOrgShare(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("shareID"))
	if err := h.shares.RevokeOrgShare(c.UserContext(), c.Params("organizationID"), id); err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

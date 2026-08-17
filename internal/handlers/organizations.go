package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type OrganizationHandler struct{ organizations *services.OrganizationService }

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,max=120"`
	Slug string `json:"slug" validate:"required,max=63"`
}

type AddMemberRequest struct {
	UserID string            `json:"userId"`
	Role   models.MemberRole `json:"role"`
}

func NewOrganizationHandler(organizations *services.OrganizationService) (*OrganizationHandler, error) {

	if organizations == nil {
		return nil, fmt.Errorf("organization service is required")
	}

	return &OrganizationHandler{organizations: organizations}, nil
}

func (h *OrganizationHandler) Create(c *fiber.Ctx) error {
	var input CreateOrganizationRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode organization request: %w", err))
	}

	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	organization, err := h.organizations.Create(c.UserContext(), userID, input.Name, input.Slug)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(organization)
}

func (h *OrganizationHandler) List(c *fiber.Ctx) error {
	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	organizations, err := h.organizations.ListForUser(c.UserContext(), userID)

	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(organizations)
}

func (h *OrganizationHandler) CheckSlug(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Query("slug"))

	if slug == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("slug is required"))
	}

	available, err := h.organizations.IsSlugAvailable(c.UserContext(), slug)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(fiber.Map{"available": available})
}

func (h *OrganizationHandler) AddMember(c *fiber.Ctx) error {
	var input AddMemberRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode member request: %w", err))
	}

	organizationID := strings.TrimSpace(c.Params("organizationID"))

	if err := h.organizations.AddMember(c.UserContext(), organizationID, input.UserID, input.Role); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func sessionUserID(c *fiber.Ctx) (string, error) {
	session, ok := c.Locals("session").(models.Session)

	if ok && session.UserID != "" {
		return session.UserID, nil
	}

	if userID, ok := c.Locals("apiKeyUserID").(string); ok && userID != "" {
		return userID, nil
	}

	return "", fmt.Errorf("authenticated session or API key is required")
}

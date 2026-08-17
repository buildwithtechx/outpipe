package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/pkg/utils"
)

type AdminHandler struct{ admin *services.AdminService }

type updateUserStatusRequest struct {
	Status models.UserStatus `json:"status" validate:"required"`
}

func NewAdminHandler(admin *services.AdminService) (*AdminHandler, error) {
	if admin == nil {
		return nil, fmt.Errorf("admin service is required")
	}
	return &AdminHandler{admin: admin}, nil
}

func (h *AdminHandler) Overview(c *fiber.Ctx) error {
	overview, err := h.admin.Overview(c.UserContext())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(overview)
}

func (h *AdminHandler) Users(c *fiber.Ctx) error {
	users, total, err := h.admin.Users(c.UserContext(), pageValue(c, "limit"), pageValue(c, "offset"))
	if err != nil {
		return writeAdminError(c, err)
	}
	return c.JSON(fiber.Map{"items": users, "total": total})
}

func (h *AdminHandler) Organizations(c *fiber.Ctx) error {
	organizations, total, err := h.admin.Organizations(c.UserContext(), pageValue(c, "limit"), pageValue(c, "offset"))
	if err != nil {
		return writeAdminError(c, err)
	}
	return c.JSON(fiber.Map{"items": organizations, "total": total})
}

func (h *AdminHandler) Tunnels(c *fiber.Ctx) error {
	tunnels, total, err := h.admin.Tunnels(c.UserContext(), pageValue(c, "limit"), pageValue(c, "offset"))
	if err != nil {
		return writeAdminError(c, err)
	}
	return c.JSON(fiber.Map{"items": tunnels, "total": total})
}

func (h *AdminHandler) Subscriptions(c *fiber.Ctx) error {
	subscriptions, total, err := h.admin.Subscriptions(c.UserContext(), pageValue(c, "limit"), pageValue(c, "offset"))
	if err != nil {
		return writeAdminError(c, err)
	}
	return c.JSON(fiber.Map{"items": subscriptions, "total": total})
}

func (h *AdminHandler) AuditLogs(c *fiber.Ctx) error {
	events, total, err := h.admin.AuditLogs(c.UserContext(), pageValue(c, "limit"), pageValue(c, "offset"))
	if err != nil {
		return writeAdminError(c, err)
	}
	return c.JSON(fiber.Map{"items": events, "total": total})
}

func writeAdminError(c *fiber.Ctx, err error) error {
	if utils.IsClientError(err) {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return writeError(c, fiber.StatusInternalServerError, err)
}

func (h *AdminHandler) Usage(c *fiber.Ctx) error {
	usage, err := h.admin.Usage(c.UserContext())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(usage)
}

func (h *AdminHandler) SetUserStatus(c *fiber.Ctx) error {
	var input updateUserStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode user status request: %w", err))
	}
	if err := h.admin.SetUserStatus(c.UserContext(), c.Params("userID"), input.Status); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func pageValue(c *fiber.Ctx, key string) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value < 0 {
		return 0
	}
	return value
}

package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type AccountHandler struct{ accounts *services.AccountService }

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"newOwnerId" validate:"required"`
}

func NewAccountHandler(accounts *services.AccountService) (*AccountHandler, error) {

	if accounts == nil {
		return nil, fmt.Errorf("account service is required")
	}

	return &AccountHandler{accounts: accounts}, nil
}

func (h *AccountHandler) Delete(c *fiber.Ctx) error {
	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	if err := h.accounts.Delete(c.UserContext(), userID); err != nil {
		return writeError(c, fiber.StatusConflict, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AccountHandler) TransferOwnership(c *fiber.Ctx) error {
	var input TransferOwnershipRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode ownership request: %w", err))
	}

	userID, err := sessionUserID(c)

	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	if err := h.accounts.TransferOwnership(c.UserContext(), c.Params("organizationID"), userID, input.NewOwnerID); err != nil {
		return writeError(c, fiber.StatusConflict, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

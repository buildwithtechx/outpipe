package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func writeSecretError(c *fiber.Ctx, err error) error {
	var input services.SecretInputError
	if errors.As(err, &input) {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	if errors.Is(err, repositories.ErrNotFound) {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return writeError(c, fiber.StatusInternalServerError, err)
}

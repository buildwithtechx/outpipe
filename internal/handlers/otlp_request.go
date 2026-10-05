package handlers

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"mime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

func otlpBody(c *fiber.Ctx) ([]byte, error) {
	typeName, _, err := mime.ParseMediaType(c.Get("Content-Type"))
	if err != nil || (typeName != "application/json" && typeName != "application/x-protobuf") {
		return nil, fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported OTLP content type")
	}
	if c.Get("Content-Encoding") == "" {
		return c.Request().Body(), nil
	}
	if !strings.EqualFold(c.Get("Content-Encoding"), "gzip") {
		return nil, fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported content encoding")
	}
	reader, err := gzip.NewReader(bytes.NewReader(c.Request().Body()))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid gzip body")
	}
	defer reader.Close()
	body, err := io.ReadAll(io.LimitReader(reader, services.MaxOTLPBodyBytes+1))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid gzip body")
	}
	if len(body) > services.MaxOTLPBodyBytes {
		return nil, fiber.NewError(fiber.StatusRequestEntityTooLarge, "OTLP payload exceeds limit")
	}
	return body, nil
}

func writeOTLPError(c *fiber.Ctx, err error) error {
	var unavailable *services.TelemetryUnavailableError
	if errors.As(err, &unavailable) {
		c.Set("Retry-After", "5")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "telemetry backend unavailable"})
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid telemetry batch"})
}

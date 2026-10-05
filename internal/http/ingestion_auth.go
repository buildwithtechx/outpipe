package http

import (
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

func ingestionRequired(keys *services.APIKeyService, auth *services.AuthService, organizations *services.OrganizationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		credential, err := apiKeyFromRequest(c, keys)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid ingestion credential"})
		}
		key := credential.Key
		if key.OrganizationID == nil || *key.OrganizationID == "" || key.ExpiresAt == nil || key.ExpiresAt.After(time.Now().Add(24*time.Hour)) || !slices.Contains(credential.Scopes, "telemetry:write") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "an organization-scoped telemetry credential expiring within 24 hours is required"})
		}
		if auth == nil || organizations == nil {
			return c.SendStatus(fiber.StatusServiceUnavailable)
		}
		if err := auth.EnsureUserActive(c.UserContext(), key.UserID); err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		if err := organizations.Authorize(c.UserContext(), *key.OrganizationID, key.UserID, models.MemberRoleMember); err != nil {
			return c.SendStatus(fiber.StatusForbidden)
		}
		if header := c.Get("X-Organization-Id"); header != "" && header != *key.OrganizationID {
			return c.SendStatus(fiber.StatusForbidden)
		}
		if query := c.Query("org_id"); query != "" && query != *key.OrganizationID {
			return c.SendStatus(fiber.StatusForbidden)
		}
		if len(c.Request().Body()) > services.MaxOTLPBodyBytes {
			return c.SendStatus(fiber.StatusRequestEntityTooLarge)
		}
		c.Locals("ingestionOrganizationID", *key.OrganizationID)
		return c.Next()
	}
}

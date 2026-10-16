package http

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

func auditRequest(audit *services.AuditService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		userID, ok := authenticatedUserID(c)

		if audit != nil && ok {
			organizationID := c.Params("organizationID")
			var organization *string

			if organizationID != "" {
				organization = &organizationID
			}

			_ = audit.Record(context.Background(), &models.AuditEvent{OrganizationID: organization, UserID: &userID, Action: c.Method() + " " + c.Path(), ResourceType: "http", ResourceID: c.Params("tunnelID"), IPAddress: requestClientIP(c), UserAgent: c.Get("User-Agent"), Metadata: `{}`})
		}

		return err
	}
}

func sessionRequired(auth *services.AuthService, apiKeys *services.APIKeyService, cookieName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimSpace(c.Cookies(cookieName))

		if raw != "" {
			session, err := auth.AuthenticateSession(c.UserContext(), raw)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
			}

			c.Locals("session", session)
			return c.Next()
		}

		credential, err := apiKeyFromRequest(c, apiKeys)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		c.Locals("apiKeyUserID", credential.Key.UserID)
		c.Locals("apiKeyCredential", credential)

		if err := auth.EnsureUserActive(c.UserContext(), credential.Key.UserID); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Next()
	}
}

func organizationRoleRequired(organizations *services.OrganizationService, required models.MemberRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := authenticatedUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authenticated session is required"})
		}

		if credential, apiKey := c.Locals("apiKeyCredential").(services.APIKeyCredential); apiKey && !apiKeyScopeAllowed(credential, required) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key scope is insufficient"})
		}

		if credential, apiKey := c.Locals("apiKeyCredential").(services.APIKeyCredential); apiKey && credential.Key.OrganizationID != nil && *credential.Key.OrganizationID != c.Params("organizationID") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key is restricted to another organization"})
		}

		if err := organizations.Authorize(c.UserContext(), c.Params("organizationID"), userID, required); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Next()
	}
}

func apiKeyFromRequest(c *fiber.Ctx, apiKeys *services.APIKeyService) (services.APIKeyCredential, error) {
	if apiKeys == nil {
		return services.APIKeyCredential{}, fmt.Errorf("authentication is unavailable")
	}

	value := strings.TrimSpace(c.Get("Authorization"))
	parts := strings.SplitN(value, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return services.APIKeyCredential{}, fmt.Errorf("session or bearer API key is required")
	}

	return apiKeys.Authenticate(c.UserContext(), strings.TrimSpace(parts[1]))
}

func agentTokenRequired(agents *services.AgentService, parameter string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if agents == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "agent authentication is unavailable"})
		}

		value := strings.TrimSpace(c.Get("Authorization"))
		parts := strings.SplitN(value, " ", 2)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "bearer agent token is required"})
		}

		agent, err := agents.Authenticate(c.UserContext(), strings.TrimSpace(parts[1]))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid agent token"})
		}

		if agent.ID != c.Params(parameter) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "agent token does not match the requested agent"})
		}

		c.Locals("agent", agent)
		return c.Next()
	}
}

func authenticatedUserID(c *fiber.Ctx) (string, bool) {
	if session, ok := c.Locals("session").(models.Session); ok && session.UserID != "" {
		return session.UserID, true
	}

	userID, ok := c.Locals("apiKeyUserID").(string)
	return userID, ok && userID != ""
}

func platformAdminRequired(auth *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session, ok := c.Locals("session").(models.Session)
		if !ok || session.UserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authenticated session is required"})
		}

		admin, err := auth.IsPlatformAdmin(c.UserContext(), session.UserID)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "platform admin access is unavailable"})
		}

		if !admin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "platform admin access is required"})
		}

		return c.Next()
	}
}

func statusPageRequired(uptime *services.UptimeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if uptime == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "status page service is unavailable"})
		}
		page, err := uptime.GetStatusPageByOrg(c.UserContext(), c.Params("organizationID"))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "status page not found"})
		}
		c.Locals("statusPage", *page)
		return c.Next()
	}
}

func internalSecretRequired(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		provided := []byte(c.Get("X-Internal-Secret"))
		if expected == "" || subtle.ConstantTimeCompare(provided, []byte(expected)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid internal secret"})
		}
		return c.Next()
	}
}

func securityHeadersMiddleware(requireTLS bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if requireTLS || c.Protocol() == "https" || c.Get("X-Forwarded-Proto") == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		return c.Next()
	}
}

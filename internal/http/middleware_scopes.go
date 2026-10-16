package http

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

func apiKeyScopeAllowed(credential services.APIKeyCredential, required models.MemberRole) bool {
	scope := "organization:read"
	if required == models.MemberRoleMember {
		scope = "organization:write"
	}
	if required == models.MemberRoleAdmin {
		scope = "organization:admin"
	}
	if required == models.MemberRoleOwner {
		scope = "organization:owner"
	}

	for _, granted := range credential.Scopes {
		if scopeAllowed(granted, scope) {
			return true
		}
	}
	return false
}

func apiKeyScopeRequired(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		credential, ok := c.Locals("apiKeyCredential").(services.APIKeyCredential)
		if !ok {
			return c.Next()
		}

		for _, granted := range credential.Scopes {
			if scopeAllowed(granted, scope) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key scope is insufficient"})
	}
}

func apiKeyResourceScopeRequired(organizations *services.OrganizationService, scope, parameter string, resolve func(context.Context, string) (string, error)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, authenticated := authenticatedUserID(c)
		if !authenticated {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authenticated session is required"})
		}

		credential, ok := c.Locals("apiKeyCredential").(services.APIKeyCredential)
		if ok && !scopeAllowedForCredential(credential, scope) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key scope is insufficient"})
		}

		organizationID, err := resolve(c.UserContext(), c.Params(parameter))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		if ok && credential.Key.OrganizationID != nil && organizationID != *credential.Key.OrganizationID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key is restricted to another organization"})
		}

		role := models.MemberRoleViewer
		if strings.HasSuffix(scope, ":write") {
			role = models.MemberRoleMember
		}
		if strings.HasPrefix(scope, "agents:") || strings.HasPrefix(scope, "domains:") {
			role = models.MemberRoleAdmin
		}

		if err := organizations.Authorize(c.UserContext(), organizationID, userID, role); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Next()
	}
}

func scopeAllowedForCredential(credential services.APIKeyCredential, required string) bool {
	for _, granted := range credential.Scopes {
		if scopeAllowed(granted, required) {
			return true
		}
	}
	return false
}

func scopeAllowed(granted, required string) bool {
	if granted == "*" || granted == required {
		return true
	}

	grantedParts := strings.SplitN(granted, ":", 2)
	requiredParts := strings.SplitN(required, ":", 2)

	if len(grantedParts) != 2 || len(requiredParts) != 2 || grantedParts[0] != requiredParts[0] {
		return false
	}

	ranks := map[string]int{"read": 1, "write": 2, "admin": 3, "owner": 4}
	return ranks[grantedParts[1]] >= ranks[requiredParts[1]]
}

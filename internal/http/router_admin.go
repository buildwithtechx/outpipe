package http

import (
	"github.com/gofiber/fiber/v2"
)

func registerAdminRoutes(protected fiber.Router, handlers Handlers) {
	admin := protected.Group("/admin", platformAdminRequired(handlers.authService))
	admin.Get("/overview", handlers.Admin.Overview)
	admin.Get("/users", handlers.Admin.Users)
	admin.Get("/users/:userID", handlers.Admin.User)
	admin.Patch("/users/:userID/status", handlers.Admin.SetUserStatus)
	admin.Get("/organizations", handlers.Admin.Organizations)
	admin.Get("/organizations/:organizationID", handlers.Admin.Organization)
	admin.Get("/tunnels", handlers.Admin.Tunnels)
	admin.Get("/subscriptions", handlers.Admin.Subscriptions)
	admin.Get("/usage", handlers.Admin.Usage)
	admin.Get("/audit-logs", handlers.Admin.AuditLogs)
	admin.Post("/tunnels/:tunnelID/revoke", handlers.Tunnels.Revoke)
}

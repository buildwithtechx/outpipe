package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type SessionCookieConfig struct {
	Name   string
	Secure bool
	Domain string
}

func SetSessionCookie(c *fiber.Ctx, cfg SessionCookieConfig, value string, expires time.Time) {
	c.Cookie(&fiber.Cookie{Name: cfg.Name, Value: value, HTTPOnly: true, Secure: cfg.Secure, SameSite: "Lax", Path: "/", Domain: cfg.Domain, Expires: expires})
}

func ClearSessionCookie(c *fiber.Ctx, cfg SessionCookieConfig) {
	c.Cookie(&fiber.Cookie{Name: cfg.Name, Value: "", HTTPOnly: true, Secure: cfg.Secure, SameSite: "Lax", Path: "/", Domain: cfg.Domain, Expires: time.Unix(0, 0)})
}

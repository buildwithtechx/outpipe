package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type OAuthHandler struct {
	oauth        *services.OAuthService
	publicAPIURL string
	dashboardURL string
	cookieName   string
	cookieSecure bool
}

func NewOAuthHandler(oauth *services.OAuthService, publicAPIURL, dashboardURL, cookieName string, cookieSecure bool) (*OAuthHandler, error) {
	if oauth == nil || strings.TrimSpace(publicAPIURL) == "" || strings.TrimSpace(dashboardURL) == "" || strings.TrimSpace(cookieName) == "" {
		return nil, fmt.Errorf("oauth service, public api url, dashboard url, and cookie name are required")
	}
	return &OAuthHandler{oauth: oauth, publicAPIURL: strings.TrimRight(publicAPIURL, "/"), dashboardURL: strings.TrimRight(dashboardURL, "/"), cookieName: cookieName, cookieSecure: cookieSecure}, nil
}

func (h *OAuthHandler) Start(c *fiber.Ctx) error {
	provider := c.Params("provider")
	returnPath, err := validatedReturnPath(c.Query("return_to"))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	redirectURI := h.publicAPIURL + "/api/v1/auth/oauth/" + provider + "/callback"
	url, err := h.oauth.Start(c.UserContext(), provider, redirectURI, returnPath)
	if err != nil {
		return c.Redirect(h.dashboardURL+"/login?error=oauth_start_failed", fiber.StatusFound)
	}
	return c.Redirect(url, fiber.StatusFound)
}

func (h *OAuthHandler) Callback(c *fiber.Ctx) error {
	raw, session, returnPath, err := h.oauth.Callback(c.UserContext(), c.Query("state"), c.Query("code"), c.Get("User-Agent"), c.IP())
	if err != nil {
		return c.Redirect(h.dashboardURL+"/login?error=oauth_failed", fiber.StatusFound)
	}
	c.Cookie(&fiber.Cookie{Name: h.cookieName, Value: raw, HTTPOnly: true, Secure: h.cookieSecure, SameSite: "Lax", Path: "/", Expires: session.ExpiresAt})
	return c.Redirect(h.dashboardURL+returnPath, fiber.StatusFound)
}

func validatedReturnPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/", nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "", fmt.Errorf("return_to must be a dashboard-relative path")
	}
	return value, nil
}

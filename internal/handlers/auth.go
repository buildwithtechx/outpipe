package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type AuthHandler struct {
	auth         *services.AuthService
	deviceLogin  *services.DeviceLoginService
	cookieName   string
	cookieSecure bool
}

type StartDeviceLoginRequest struct {
	IPAddress string `json:"ipAddress"`
}

type CompleteDeviceLoginRequest struct {
	Code string `json:"code"`
}

func NewAuthHandler(auth *services.AuthService, deviceLogin *services.DeviceLoginService, cookieName string, cookieSecure bool) (*AuthHandler, error) {
	if auth == nil || deviceLogin == nil || strings.TrimSpace(cookieName) == "" {
		return nil, fmt.Errorf("auth services and cookie name are required")
	}
	return &AuthHandler{auth: auth, deviceLogin: deviceLogin, cookieName: cookieName, cookieSecure: cookieSecure}, nil
}

func (h *AuthHandler) StartDeviceLogin(c *fiber.Ctx) error {
	var input StartDeviceLoginRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode device login request: %w", err))
		}
	}
	code, login, err := h.deviceLogin.Start(c.UserContext(), input.IPAddress)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": login.ID, "code": code, "expiresAt": login.ExpiresAt})
}

func (h *AuthHandler) CompleteDeviceLogin(c *fiber.Ctx) error {
	var input CompleteDeviceLoginRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode device completion request: %w", err))
	}
	userID, err := sessionUserID(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	token, err := h.deviceLogin.Complete(c.UserContext(), input.Code, userID)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	return c.JSON(fiber.Map{"token": token})
}

func (h *AuthHandler) PollDeviceLogin(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("device code is required"))
	}
	token, complete, err := h.deviceLogin.Poll(c.UserContext(), code)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	if !complete {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "pending"})
	}
	return c.JSON(fiber.Map{"status": "complete", "token": token})
}

func (h *AuthHandler) Session(c *fiber.Ctx) error {
	raw := c.Cookies(h.cookieName)
	session, err := h.auth.AuthenticateSession(c.UserContext(), raw)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	isPlatformAdmin, err := h.auth.IsPlatformAdmin(c.UserContext(), session.UserID)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, fmt.Errorf("check platform admin access: %w", err))
	}
	user, err := h.auth.CurrentUser(c.UserContext(), session.UserID)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	return c.JSON(fiber.Map{"id": session.ID, "userId": session.UserID, "expiresAt": session.ExpiresAt, "lastSeenAt": session.LastSeenAt, "isPlatformAdmin": isPlatformAdmin, "user": user})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	raw := c.Cookies(h.cookieName)
	session, err := h.auth.AuthenticateSession(c.UserContext(), raw)
	if err == nil {
		err = h.auth.RevokeSession(c.UserContext(), session.ID)
	}
	c.Cookie(&fiber.Cookie{Name: h.cookieName, Value: "", Expires: time.Unix(0, 0), HTTPOnly: true, Secure: h.cookieSecure, SameSite: "Lax"})
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

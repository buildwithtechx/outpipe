package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

var (
	errCertificateIdentityRequired = errors.New("hostname or tunnel id is required")
	errCertificateUnavailable      = errors.New("certificate is unavailable")
)

func (h *DomainHandler) ServeCertificate(c *fiber.Ctx) error {
	hostname := strings.ToLower(strings.TrimSpace(c.Query("hostname")))
	tunnelID := strings.TrimSpace(c.Query("tunnelId"))
	if hostname == "" && tunnelID == "" {
		return writeError(c, fiber.StatusBadRequest, errCertificateIdentityRequired)
	}
	if tunnelID != "" && h.tunnels != nil {
		tunnel, err := h.tunnels.Find(c.UserContext(), tunnelID)
		if err != nil {
			return writeError(c, fiber.StatusNotFound, err)
		}
		if tunnel.RevokedAt != nil {
			return writeError(c, fiber.StatusNotFound, errCertificateUnavailable)
		}
		if hostname == "" {
			hostname = strings.ToLower(strings.TrimSpace(tunnel.PublicHostname))
		} else if !strings.EqualFold(strings.TrimSpace(tunnel.PublicHostname), hostname) {
			return writeError(c, fiber.StatusForbidden, errCertificateUnavailable)
		}
		certificate, err := h.domains.ServeCertificate(c.UserContext(), tunnel.OrganizationID, hostname)
		if err != nil {
			return certificateError(c, err)
		}
		return serveCertificateJSON(c, certificate)
	}
	if hostname == "" {
		return writeError(c, fiber.StatusBadRequest, errCertificateIdentityRequired)
	}
	domain, err := h.domains.FindByHostname(c.UserContext(), hostname)
	if err != nil {
		if wildcard := parentCertificateHostname(hostname); wildcard != "" {
			if parent, parentErr := h.domains.FindByHostname(c.UserContext(), wildcard); parentErr == nil {
				domain = parent
				err = nil
			}
		}
	}
	if err != nil {
		return writeError(c, fiber.StatusNotFound, errCertificateUnavailable)
	}
	certificate, err := h.domains.ServeCertificate(c.UserContext(), domain.OrganizationID, hostname)
	if err != nil {
		return certificateError(c, err)
	}
	return serveCertificateJSON(c, certificate)
}

func certificateError(c *fiber.Ctx, err error) error {
	if err == services.ErrCertificateUnavailable {
		return writeError(c, fiber.StatusNotFound, err)
	}
	return writeError(c, fiber.StatusBadRequest, err)
}

func serveCertificateJSON(c *fiber.Ctx, certificate services.DomainCertificate) error {
	return c.JSON(fiber.Map{
		"hostname":    certificate.Hostname,
		"certificate": string(certificate.ChainPEM),
		"key":         string(certificate.KeyPEM),
		"expiresAt":   certificate.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"wildcardFor": certificate.WildcardFor,
	})
}

func parentCertificateHostname(hostname string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(hostname)), ".")
	if len(parts) < 3 {
		return ""
	}
	return "*." + strings.Join(parts[1:], ".")
}

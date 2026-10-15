package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
)

type DomainHandler struct {
	domains *services.DomainService
	tunnels *services.TunnelService
}

func (h *DomainHandler) SetTunnels(tunnels *services.TunnelService) {
	h.tunnels = tunnels
}

func (h *DomainHandler) OrganizationID(ctx context.Context, id string) (string, error) {
	domain, err := h.domains.Find(ctx, id)
	return domain.OrganizationID, err
}

type CreateDomainRequest struct {
	Hostname           string  `json:"hostname"`
	VerificationMethod string  `json:"verificationMethod"`
	TunnelID           *string `json:"tunnelId"`
	DNSProvider        string  `json:"dnsProvider"`
}

type VerifyDomainRequest struct {
	Token string `json:"token"`
}

func NewDomainHandler(domains *services.DomainService) (*DomainHandler, error) {

	if domains == nil {
		return nil, fmt.Errorf("domain service is required")
	}

	return &DomainHandler{domains: domains}, nil
}

func (h *DomainHandler) Create(c *fiber.Ctx) error {
	var input CreateDomainRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode domain request: %w", err))
	}

	token, domain, err := h.domains.CreateWithProvider(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), input.Hostname, input.VerificationMethod, input.TunnelID, input.DNSProvider)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"domain": domain, "verificationToken": token})
}

type CreateDomainBundleRequest struct {
	Hostname    string  `json:"hostname"`
	TunnelID    *string `json:"tunnelId"`
	DNSProvider string  `json:"dnsProvider"`
	Wildcard    bool    `json:"wildcard"`
	Delegated   bool    `json:"delegated"`
}

func (h *DomainHandler) CreateBundle(c *fiber.Ctx) error {
	var input CreateDomainBundleRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode domain bundle request: %w", err))
	}

	method := services.DomainMethodDNSAutomation
	if input.Delegated {
		method = services.DomainMethodDNSDelegation
	}

	domains, err := h.domains.CreateBundle(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), input.Hostname, method, input.TunnelID, input.DNSProvider, input.Wildcard)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(domains)
}

func (h *DomainHandler) List(c *fiber.Ctx) error {
	domains, err := h.domains.ListByOrganization(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	return c.JSON(domains)
}

func (h *DomainHandler) Verify(c *fiber.Ctx) error {
	var input VerifyDomainRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode domain verification request: %w", err))
	}

	if err := h.domains.Verify(c.UserContext(), c.Params("domainID"), input.Token); err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *DomainHandler) Delete(c *fiber.Ctx) error {
	if err := h.domains.Delete(c.UserContext(), c.Params("domainID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *DomainHandler) Revoke(c *fiber.Ctx) error {
	if err := h.domains.RevokeDomainCertificate(c.UserContext(), c.Params("domainID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type StoreDNSProviderRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
	Label    string `json:"label"`
}

func (h *DomainHandler) StoreDNSProvider(c *fiber.Ctx) error {
	var input StoreDNSProviderRequest

	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode dns provider request: %w", err))
	}

	credential, err := h.domains.StoreDNSCredential(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), input.Provider, input.Token, input.Label)

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(credential)
}

func (h *DomainHandler) ListDNSProviders(c *fiber.Ctx) error {
	credentials, err := h.domains.ListDNSCredentials(c.UserContext(), strings.TrimSpace(c.Params("organizationID")))

	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(credentials)
}

func (h *DomainHandler) DeleteDNSProvider(c *fiber.Ctx) error {
	if err := h.domains.DeleteDNSCredential(c.UserContext(), strings.TrimSpace(c.Params("organizationID")), c.Params("credentialID")); err != nil {
		return writeError(c, fiber.StatusNotFound, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

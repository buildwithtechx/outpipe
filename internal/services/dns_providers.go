package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/infra/dns"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type dnsSecretProtector interface {
	Seal(string) (string, error)
	Open(string) (string, error)
}

type DNSCredentialService struct {
	repository repositories.DNSProviderRepository
	protector  dnsSecretProtector
	now        func() time.Time
}

func NewDNSCredentialService(repository repositories.DNSProviderRepository) (*DNSCredentialService, error) {
	if repository == nil {
		return nil, fmt.Errorf("dns provider repository is required")
	}
	return &DNSCredentialService{repository: repository, now: time.Now}, nil
}

func (s *DNSCredentialService) SetProtector(protector dnsSecretProtector) {
	s.protector = protector
}

func supportedDNSProvider(provider string) bool {
	return strings.ToLower(strings.TrimSpace(provider)) == "cloudflare"
}

func SolverFor(provider, token string) (dns.ChallengeSolver, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "cloudflare":
		return dns.NewCloudflareProvider(token, nil)
	default:
		return nil, fmt.Errorf("unsupported DNS provider %q", provider)
	}
}

func (s *DNSCredentialService) Store(ctx context.Context, organizationID, provider, token, label string) (models.DNSProviderCredential, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	token = strings.TrimSpace(token)
	if organizationID == "" || !supportedDNSProvider(provider) || token == "" || len(token) > 1024 {
		return models.DNSProviderCredential{}, fmt.Errorf("organization, supported provider, and token are required")
	}
	if len(label) > 80 {
		return models.DNSProviderCredential{}, fmt.Errorf("dns credential label exceeds limit")
	}
	if s.protector == nil {
		return models.DNSProviderCredential{}, fmt.Errorf("dns credential protector is required")
	}
	ciphertext, err := s.protector.Seal(token)
	if err != nil {
		return models.DNSProviderCredential{}, fmt.Errorf("encrypt dns credential: %w", err)
	}
	suffix := token
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	credential := models.DNSProviderCredential{OrganizationID: organizationID, Provider: provider, Ciphertext: ciphertext, KeyVersion: 1, Label: strings.TrimSpace(label), LastFour: suffix}
	if err := s.repository.Upsert(ctx, &credential); err != nil {
		return models.DNSProviderCredential{}, err
	}
	return credential, nil
}

func (s *DNSCredentialService) List(ctx context.Context, organizationID string) ([]models.DNSProviderCredential, error) {
	if strings.TrimSpace(organizationID) == "" {
		return nil, fmt.Errorf("organization is required")
	}
	return s.repository.FindByOrganization(ctx, organizationID)
}

func (s *DNSCredentialService) Delete(ctx context.Context, organizationID, id string) error {
	credential, err := s.repository.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if credential.OrganizationID != organizationID {
		return repositories.ErrNotFound
	}
	return s.repository.Delete(ctx, credential.ID)
}

func (s *DNSCredentialService) Open(ctx context.Context, organizationID, provider string) (string, error) {
	if s.protector == nil {
		return "", fmt.Errorf("dns credential protector is required")
	}
	credentials, err := s.repository.FindByOrganization(ctx, organizationID)
	if err != nil {
		return "", err
	}
	for _, credential := range credentials {
		if credential.Provider != provider {
			continue
		}
		token, err := s.protector.Open(credential.Ciphertext)
		if err != nil {
			return "", fmt.Errorf("decrypt dns credential: %w", err)
		}
		return token, nil
	}
	return "", fmt.Errorf("dns provider %q is not connected", provider)
}

func (s *DNSCredentialService) TouchUsed(ctx context.Context, organizationID string) {
	credentials, err := s.repository.FindByOrganization(ctx, organizationID)
	if err != nil || len(credentials) == 0 {
		return
	}
	now := s.now()
	credentials[0].LastUsedAt = &now
	_ = s.repository.Upsert(ctx, &credentials[0])
}

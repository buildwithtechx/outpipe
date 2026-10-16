package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/models"
)

var ErrCertificateUnavailable = fmt.Errorf("certificate is unavailable")

var errDomainCertificateUnavailable = ErrCertificateUnavailable

func (s *DomainService) ServeCertificate(ctx context.Context, organizationID, hostname string) (DomainCertificate, error) {
	host := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))
	if host == "" || strings.Contains(host, "*") {
		return DomainCertificate{}, fmt.Errorf("hostname is required")
	}
	if strings.TrimSpace(organizationID) == "" {
		return DomainCertificate{}, fmt.Errorf("organization is required")
	}
	if s.certCacheDir == "" {
		return DomainCertificate{}, fmt.Errorf("certificate cache is not configured")
	}
	domain, wildcardFor, err := s.certificateDomainForHost(ctx, host)
	if err != nil {
		return DomainCertificate{}, err
	}
	if domain.Status == models.DomainStatusRevoked {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	if domain.Status != models.DomainStatusActive || domain.CertificateStatus != "ready" {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	if domain.CertificateExpiresAt == nil || !s.now().Before(*domain.CertificateExpiresAt) {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	if domain.OrganizationID != organizationID {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	certFile, keyFile := certificates.HostCertificatePaths(s.certCacheDir, domain.Hostname)
	chainPEM, err := os.ReadFile(certFile)
	if err != nil {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	if len(chainPEM) == 0 || len(keyPEM) == 0 {
		return DomainCertificate{}, errDomainCertificateUnavailable
	}
	return DomainCertificate{Hostname: domain.Hostname, ChainPEM: chainPEM, KeyPEM: keyPEM, ExpiresAt: *domain.CertificateExpiresAt, WildcardFor: wildcardFor}, nil
}

func (s *DomainService) certificateDomainForHost(ctx context.Context, host string) (models.Domain, string, error) {
	if domain, err := s.domains.FindByHostname(ctx, host); err == nil {
		return domain, "", nil
	}
	if parent := parentDomain(host); parent != "" {
		if wildcard, err := s.domains.FindByHostname(ctx, "*."+parent); err == nil {
			return wildcard, host, nil
		}
	}
	return models.Domain{}, "", errDomainCertificateUnavailable
}

func parentDomain(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}

func (s *DomainService) RevokeDomainCertificate(ctx context.Context, id string) error {
	domain, err := s.domains.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("find domain: %w", err)
	}
	now := s.now()
	domain.Status = models.DomainStatusRevoked
	domain.CertificateStatus = "revoked"
	domain.CertificateExpiresAt = &now
	if err := s.domains.Update(ctx, &domain); err != nil {
		return fmt.Errorf("revoke domain certificate: %w", err)
	}
	s.removeCertificateFiles(domain.Hostname)
	return nil
}

func (s *DomainService) removeCertificateFiles(hostname string) {
	if s.certCacheDir == "" {
		return
	}
	certFile, keyFile := certificates.HostCertificatePaths(s.certCacheDir, hostname)
	_ = os.Remove(certFile)
	_ = os.Remove(keyFile)
}

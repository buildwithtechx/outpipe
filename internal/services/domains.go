package services

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type DomainService struct {
	domains repositories.DomainRepository
	billing *BillingService
	now     func() time.Time
	issuer  CertificateIssuer
	dns     DNSProvider
	dnsTTL  int
}

type CertificateIssuer interface {
	Issue(context.Context, string) (time.Time, error)
}

type DNSProvider interface {
	UpsertTXT(context.Context, string, string, int) error
}

func (s *DomainService) SetBilling(billing *BillingService)            { s.billing = billing }
func (s *DomainService) SetCertificateIssuer(issuer CertificateIssuer) { s.issuer = issuer }
func (s *DomainService) SetDNSTTL(ttl int) {
	if ttl > 0 {
		s.dnsTTL = ttl
	}
}
func (s *DomainService) SetDNSProvider(provider DNSProvider) { s.dns = provider }

func (s *DomainService) Find(ctx context.Context, id string) (models.Domain, error) {
	domain, err := s.domains.FindByID(ctx, id)
	if err != nil {
		return models.Domain{}, fmt.Errorf("find domain: %w", err)
	}
	return domain, nil
}

func NewDomainService(domains repositories.DomainRepository) (*DomainService, error) {
	if domains == nil {
		return nil, fmt.Errorf("domain repository is required")
	}
	return &DomainService{domains: domains, now: time.Now, dnsTTL: 120}, nil
}

func (s *DomainService) Create(ctx context.Context, organizationID, hostname, method string, tunnelID *string) (string, models.Domain, error) {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	method = strings.ToLower(strings.TrimSpace(method))
	if organizationID == "" || !validHostname(hostname) || method == "" {
		return "", models.Domain{}, fmt.Errorf("organization, hostname, and verification method are required")
	}
	if s.billing != nil {
		plan, _, err := s.billing.Entitlements(ctx, organizationID)
		if err != nil {
			return "", models.Domain{}, fmt.Errorf("check domain entitlement: %w", err)
		}
		if plan.MaxDomains == 0 {
			return "", models.Domain{}, fmt.Errorf("custom domains are not available on this plan")
		}
		count, err := s.domains.CountByOrganization(ctx, organizationID)
		if err != nil {
			return "", models.Domain{}, fmt.Errorf("count organization domains: %w", err)
		}
		if count >= int64(plan.MaxDomains) {
			return "", models.Domain{}, fmt.Errorf("organization domain limit reached")
		}
	}
	raw, err := auth.NewToken("cdv", 24)
	if err != nil {
		return "", models.Domain{}, err
	}
	domain := models.Domain{OrganizationID: organizationID, TunnelID: tunnelID, Hostname: hostname, Status: models.DomainStatusPending, VerificationMethod: method, VerificationToken: auth.HashToken(raw), CertificateStatus: "pending"}
	if method == "dns" && s.dns != nil {
		if err := s.dns.UpsertTXT(ctx, "_outpipe-challenge."+hostname, raw, s.dnsTTL); err != nil {
			return "", models.Domain{}, fmt.Errorf("create dns challenge: %w", err)
		}
	}
	if err := s.domains.Create(ctx, &domain); err != nil {
		return "", models.Domain{}, fmt.Errorf("create domain: %w", err)
	}
	return raw, domain, nil
}

func (s *DomainService) Verify(ctx context.Context, id, token string) error {
	domain, err := s.domains.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find domain: %w", err)
	}
	if !auth.EqualHash(domain.VerificationToken, token) {
		return fmt.Errorf("invalid domain verification token")
	}
	now := s.now()
	domain.Status = models.DomainStatusVerified
	domain.VerifiedAt = &now
	if s.issuer != nil {
		expires, err := s.issuer.Issue(ctx, domain.Hostname)
		if err != nil {
			return fmt.Errorf("issue domain certificate: %w", err)
		}
		domain.Status = models.DomainStatusActive
		domain.CertificateStatus = "ready"
		domain.CertificateExpiresAt = &expires
	}
	if err := s.domains.Update(ctx, &domain); err != nil {
		return fmt.Errorf("verify domain: %w", err)
	}
	return nil
}

func validHostname(hostname string) bool {
	if len(hostname) > 253 || !strings.Contains(hostname, ".") || strings.Contains(hostname, "..") || net.ParseIP(hostname) != nil {
		return false
	}
	for _, label := range strings.Split(hostname, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}

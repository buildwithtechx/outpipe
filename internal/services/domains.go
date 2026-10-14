package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/infra/dns"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/validation"
)

const (
	DomainMethodDNS           = "dns"
	DomainMethodHTTP          = "http"
	DomainMethodDNSAutomation = "dns-automation"
)

type DomainService struct {
	domains        repositories.DomainRepository
	billing        *BillingService
	now            func() time.Time
	issuer         CertificateIssuer
	dnsCredentials *DNSCredentialService
	dnsIssuer      DNS01Issuer
	dnsSolver      func(string, string) (dns.ChallengeSolver, error)
	httpClient     *http.Client
	certCacheDir   string
}

type DomainCertificate struct {
	Hostname    string
	ChainPEM    []byte
	KeyPEM      []byte
	ExpiresAt   time.Time
	WildcardFor string
}

type CertificateIssuer interface {
	Issue(context.Context, string) (time.Time, error)
}

type DNS01Issuer interface {
	IssueDNS01(context.Context, string, dns.ChallengeSolver) ([]byte, []byte, time.Time, error)
}

func (s *DomainService) SetBilling(billing *BillingService)                  { s.billing = billing }
func (s *DomainService) SetCertificateIssuer(issuer CertificateIssuer)       { s.issuer = issuer }
func (s *DomainService) SetCertificateCacheDir(dir string) {
	s.certCacheDir = strings.TrimSpace(dir)
}
func (s *DomainService) SetDNSCredentials(credentials *DNSCredentialService) { s.dnsCredentials = credentials }
func (s *DomainService) SetDNS01Issuer(issuer DNS01Issuer)                    { s.dnsIssuer = issuer }
func (s *DomainService) SetDNSProtector(protector dnsSecretProtector) {
	if s.dnsCredentials != nil {
		s.dnsCredentials.SetProtector(protector)
	}
}

func (s *DomainService) Find(ctx context.Context, id string) (models.Domain, error) {
	domain, err := s.domains.FindByID(ctx, id)

	if err != nil {
		return models.Domain{}, fmt.Errorf("find domain: %w", err)
	}

	return domain, nil
}

func (s *DomainService) FindByHostname(ctx context.Context, hostname string) (models.Domain, error) {
	host := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))

	if host == "" {
		return models.Domain{}, fmt.Errorf("hostname is required")
	}

	domain, err := s.domains.FindByHostname(ctx, host)

	if err != nil {
		return models.Domain{}, fmt.Errorf("find domain: %w", err)
	}

	return domain, nil
}

func (s *DomainService) ListByOrganization(ctx context.Context, organizationID string) ([]models.Domain, error) {
	if strings.TrimSpace(organizationID) == "" {
		return nil, fmt.Errorf("organization is required")
	}
	domains, err := s.domains.FindByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization domains: %w", err)
	}
	return domains, nil
}

func (s *DomainService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("domain is required")
	}
	domain, err := s.domains.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("find domain: %w", err)
	}
	if err := s.domains.Delete(ctx, strings.TrimSpace(id)); err != nil {
		return err
	}
	s.removeCertificateFiles(domain.Hostname)
	return nil
}

func (s *DomainService) StoreDNSCredential(ctx context.Context, organizationID, provider, token, label string) (models.DNSProviderCredential, error) {
	if s.dnsCredentials == nil {
		return models.DNSProviderCredential{}, fmt.Errorf("dns credentials are not configured")
	}

	return s.dnsCredentials.Store(ctx, organizationID, provider, token, label)
}

func (s *DomainService) ListDNSCredentials(ctx context.Context, organizationID string) ([]models.DNSProviderCredential, error) {
	if s.dnsCredentials == nil {
		return nil, fmt.Errorf("dns credentials are not configured")
	}

	return s.dnsCredentials.List(ctx, organizationID)
}

func (s *DomainService) DeleteDNSCredential(ctx context.Context, organizationID, id string) error {
	if s.dnsCredentials == nil {
		return fmt.Errorf("dns credentials are not configured")
	}

	return s.dnsCredentials.Delete(ctx, organizationID, id)
}

func NewDomainService(domains repositories.DomainRepository) (*DomainService, error) {

	if domains == nil {
		return nil, fmt.Errorf("domain repository is required")
	}

	return &DomainService{domains: domains, now: time.Now, dnsSolver: SolverFor, httpClient: validation.NewSafeHTTPClient(5 * time.Second)}, nil
}

func (s *DomainService) Create(ctx context.Context, organizationID, hostname, method string, tunnelID *string) (string, models.Domain, error) {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	method = strings.ToLower(strings.TrimSpace(method))

	if organizationID == "" || !validHostname(hostname) || (method != DomainMethodDNS && method != DomainMethodHTTP && method != DomainMethodDNSAutomation) {
		return "", models.Domain{}, fmt.Errorf("organization, hostname, and verification method are required")
	}

	if strings.HasPrefix(hostname, "*.") && method != DomainMethodDNSAutomation {
		return "", models.Domain{}, fmt.Errorf("wildcard domains require dns-automation verification")
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

	if strings.HasPrefix(domain.Hostname, "*.") && domain.VerificationMethod != DomainMethodDNSAutomation {
		return fmt.Errorf("wildcard domains require dns-automation verification")
	}

	switch domain.VerificationMethod {
	case DomainMethodDNSAutomation:
		return s.verifyAutomated(ctx, &domain)
	case DomainMethodDNS, DomainMethodHTTP:
		return s.verifyManual(ctx, &domain, token)
	default:
		return fmt.Errorf("unsupported verification method %q", domain.VerificationMethod)
	}
}

func (s *DomainService) verifyManual(ctx context.Context, domain *models.Domain, token string) error {
	if err := validation.ValidateSafeTarget(strings.TrimPrefix(domain.Hostname, "*.")); err != nil {
		return fmt.Errorf("domain verification target is invalid: %w", err)
	}

	verified := false

	switch domain.VerificationMethod {
	case DomainMethodDNS:
		txts, err := net.DefaultResolver.LookupTXT(ctx, "_outpipe-challenge."+domain.Hostname)

		if err == nil {
			for _, txt := range txts {
				if auth.EqualHash(domain.VerificationToken, strings.TrimSpace(txt)) || txt == token {
					verified = true
					break
				}
			}
		}
	case DomainMethodHTTP:
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+domain.Hostname+"/.well-known/outpipe-challenge", nil)

		if err == nil {
			response, err := s.httpClient.Do(request)

			if err == nil {
				body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
				_ = response.Body.Close()
				content := strings.TrimSpace(string(body))

				if auth.EqualHash(domain.VerificationToken, content) || content == token {
					verified = true
				}
			}
		}
	}

	if !verified {
		return fmt.Errorf("custom domain verification failed for %s", domain.Hostname)
	}

	return s.markVerified(ctx, domain)
}

func (s *DomainService) verifyAutomated(ctx context.Context, domain *models.Domain) error {
	if s.dnsCredentials == nil || s.dnsIssuer == nil {
		return fmt.Errorf("automated verification is not configured")
	}

	solver, err := s.dnsSolverForOrganization(ctx, domain.OrganizationID)

	if err != nil {
		return err
	}

	_, _, expires, err := s.dnsIssuer.IssueDNS01(ctx, domain.Hostname, solver)

	if err != nil {
		return fmt.Errorf("issue automated certificate: %w", err)
	}

	s.dnsCredentials.TouchUsed(ctx, domain.OrganizationID)

	now := s.now()
	domain.Status = models.DomainStatusActive
	domain.VerifiedAt = &now
	domain.CertificateStatus = "ready"
	domain.CertificateExpiresAt = &expires

	if err := s.domains.Update(ctx, domain); err != nil {
		return fmt.Errorf("verify domain: %w", err)
	}

	return nil
}

func (s *DomainService) markVerified(ctx context.Context, domain *models.Domain) error {
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

	if err := s.domains.Update(ctx, domain); err != nil {
		return fmt.Errorf("verify domain: %w", err)
	}

	return nil
}

func (s *DomainService) RenewDueCertificates(ctx context.Context, within time.Duration) (int, error) {
	before := s.now().Add(within)
	domains, err := s.domains.FindExpiringCertificates(ctx, before)

	if err != nil {
		return 0, err
	}

	renewed := 0
	var failures error

	for i := range domains {
		domain := domains[i]

		var expires time.Time
		var err error

		if domain.VerificationMethod == DomainMethodDNSAutomation {
			expires, err = s.renewAutomated(ctx, &domain)
		} else if s.issuer != nil {
			expires, err = s.issuer.Issue(ctx, domain.Hostname)
		} else {
			continue
		}

		if err != nil {
			failures = errors.Join(failures, fmt.Errorf("renew %s: %w", domain.Hostname, err))
			continue
		}

		domain.CertificateExpiresAt = &expires
		domain.CertificateStatus = "ready"

		if err := s.domains.Update(ctx, &domain); err != nil {
			failures = errors.Join(failures, fmt.Errorf("record %s renewal: %w", domain.Hostname, err))
			continue
		}

		renewed++
	}

	return renewed, failures
}

func (s *DomainService) dnsSolverForOrganization(ctx context.Context, organizationID string) (dns.ChallengeSolver, error) {
	if s.dnsCredentials == nil {
		return nil, fmt.Errorf("dns credentials are not configured")
	}

	factory := s.dnsSolver

	if factory == nil {
		factory = SolverFor
	}

	credentials, err := s.dnsCredentials.List(ctx, organizationID)

	if err != nil {
		return nil, err
	}

	if len(credentials) != 1 {
		return nil, fmt.Errorf("connect exactly one DNS provider to use automated verification")
	}

	token, err := s.dnsCredentials.Open(ctx, organizationID, credentials[0].Provider)

	if err != nil {
		return nil, err
	}

	return factory(credentials[0].Provider, token)
}

func (s *DomainService) renewAutomated(ctx context.Context, domain *models.Domain) (time.Time, error) {
	if s.dnsCredentials == nil || s.dnsIssuer == nil {
		return time.Time{}, fmt.Errorf("automated renewal is not configured")
	}

	solver, err := s.dnsSolverForOrganization(ctx, domain.OrganizationID)

	if err != nil {
		return time.Time{}, err
	}

	_, _, expires, err := s.dnsIssuer.IssueDNS01(ctx, domain.Hostname, solver)

	if err != nil {
		return time.Time{}, fmt.Errorf("issue automated renewal: %w", err)
	}

	s.dnsCredentials.TouchUsed(ctx, domain.OrganizationID)

	return expires, nil
}

func validHostname(hostname string) bool {
	if strings.HasPrefix(hostname, "*.") {
		hostname = hostname[2:]
	}

	if len(hostname) > 253 || !strings.Contains(hostname, ".") || strings.Contains(hostname, "..") || strings.Contains(hostname, "*") || net.ParseIP(hostname) != nil {
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

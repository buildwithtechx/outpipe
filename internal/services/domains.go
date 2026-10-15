package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	DomainMethodDNSDelegation = "dns-delegation"
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
	challengeZone  string
	challenge      dns.ChallengeSolver
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
	IssueBundleDNS01(context.Context, []string, dns.ChallengeSolver) ([]byte, []byte, time.Time, error)
	IssueDelegatedDNS01(context.Context, string, string, dns.ChallengeSolver) ([]byte, []byte, time.Time, error)
	IssueBundleDelegatedDNS01(context.Context, []string, string, dns.ChallengeSolver) ([]byte, []byte, time.Time, error)
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

func (s *DomainService) SetChallengeSolver(zone string, solver dns.ChallengeSolver) {
	s.challengeZone = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(zone, ".")))
	s.challenge = solver
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
	return s.CreateWithProvider(ctx, organizationID, hostname, method, tunnelID, "")
}

func (s *DomainService) CreateWithProvider(ctx context.Context, organizationID, hostname, method string, tunnelID *string, dnsProvider string) (string, models.Domain, error) {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	method = strings.ToLower(strings.TrimSpace(method))
	dnsProvider = strings.ToLower(strings.TrimSpace(dnsProvider))

	if organizationID == "" || !validHostname(hostname) || (method != DomainMethodDNS && method != DomainMethodHTTP && method != DomainMethodDNSAutomation && method != DomainMethodDNSDelegation) {
		return "", models.Domain{}, fmt.Errorf("organization, hostname, and verification method are required")
	}

	if strings.HasPrefix(hostname, "*.") && method != DomainMethodDNSAutomation && method != DomainMethodDNSDelegation {
		return "", models.Domain{}, fmt.Errorf("wildcard domains require dns-automation verification")
	}

	if dnsProvider != "" && !supportedDNSProvider(dnsProvider) {
		return "", models.Domain{}, fmt.Errorf("unsupported DNS provider %q", dnsProvider)
	}

	if method == DomainMethodDNSDelegation && (s.challengeZone == "" || s.challenge == nil) {
		return "", models.Domain{}, fmt.Errorf("delegated verification is not configured")
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

	domain := models.Domain{OrganizationID: organizationID, TunnelID: tunnelID, Hostname: hostname, Status: models.DomainStatusPending, VerificationMethod: method, VerificationToken: auth.HashToken(raw), CertificateStatus: "pending", DNSProvider: dnsProvider}

	if method == DomainMethodDNSDelegation {
		target, err := s.delegationTarget()
		if err != nil {
			return "", models.Domain{}, err
		}
		domain.DelegationTarget = target
	}

	if err := s.domains.Create(ctx, &domain); err != nil {
		return "", models.Domain{}, fmt.Errorf("create domain: %w", err)
	}

	return raw, domain, nil
}

func (s *DomainService) delegationTarget() (string, error) {
	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		return "", fmt.Errorf("generate delegation target: %w", err)
	}
	return hex.EncodeToString(randomID) + "." + s.challengeZone, nil
}

func (s *DomainService) CreateBundle(ctx context.Context, organizationID, apex, method string, tunnelID *string, dnsProvider string, wildcard bool) ([]models.Domain, error) {
	apex = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(apex, ".")))
	if organizationID == "" || !validHostname(apex) || strings.HasPrefix(apex, "*.") {
		return nil, fmt.Errorf("bundle registration needs a bare apex hostname")
	}
	if method != DomainMethodDNSAutomation && method != DomainMethodDNSDelegation {
		return nil, fmt.Errorf("bundles require dns-automation verification")
	}
	hostnames := []string{apex}
	if wildcard {
		hostnames = append(hostnames, "*."+apex)
	}
	group := ""
	if len(hostnames) > 1 {
		randomID := make([]byte, 16)
		if _, err := rand.Read(randomID); err != nil {
			return nil, fmt.Errorf("generate bundle group: %w", err)
		}
		group = hex.EncodeToString(randomID)
	}
	target := ""
	if method == DomainMethodDNSDelegation {
		if s.challengeZone == "" || s.challenge == nil {
			return nil, fmt.Errorf("delegated verification is not configured")
		}
		var err error
		target, err = s.delegationTarget()
		if err != nil {
			return nil, err
		}
	}
	created := make([]models.Domain, 0, len(hostnames))
	for _, hostname := range hostnames {
		_, domain, err := s.createBundleRow(ctx, organizationID, hostname, method, tunnelID, dnsProvider, group, target)
		if err != nil {
			return nil, err
		}
		created = append(created, domain)
	}
	return created, nil
}

func (s *DomainService) createBundleRow(ctx context.Context, organizationID, hostname, method string, tunnelID *string, dnsProvider, group, target string) (string, models.Domain, error) {
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
	domain := models.Domain{OrganizationID: organizationID, TunnelID: tunnelID, Hostname: hostname, Status: models.DomainStatusPending, VerificationMethod: method, VerificationToken: auth.HashToken(raw), CertificateStatus: "pending", DNSProvider: dnsProvider, BundleGroup: group, DelegationTarget: target}
	if method == DomainMethodDNSDelegation && target == "" {
		return "", models.Domain{}, fmt.Errorf("delegated verification needs a challenge target")
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

	if strings.HasPrefix(domain.Hostname, "*.") && domain.VerificationMethod != DomainMethodDNSAutomation && domain.VerificationMethod != DomainMethodDNSDelegation {
		return fmt.Errorf("wildcard domains require dns-automation verification")
	}

	switch domain.VerificationMethod {
	case DomainMethodDNSAutomation, DomainMethodDNSDelegation:
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
	if domain.DelegationTarget != "" {
		if err := s.assertDelegation(ctx, domain); err != nil {
			return err
		}
	}
	expires, err := s.issueAutomation(ctx, domain, []string{domain.Hostname})
	if err != nil {
		return err
	}

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

func (s *DomainService) assertDelegation(ctx context.Context, domain *models.Domain) error {
	canonical, err := net.LookupCNAME("_acme-challenge." + strings.TrimPrefix(domain.Hostname, "*."))
	if err != nil {
		return fmt.Errorf("delegation record is not published: %w", err)
	}
	expected := strings.TrimSuffix(domain.DelegationTarget, ".") + "."
	if !strings.EqualFold(strings.TrimSpace(canonical), expected) {
		return fmt.Errorf("delegation record does not point at the challenge zone")
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
	renewedGroups := map[string]bool{}

	for i := range domains {
		domain := domains[i]

		if domain.BundleGroup != "" {
			if renewedGroups[domain.BundleGroup] {
				continue
			}
			renewedGroups[domain.BundleGroup] = true
			if count, err := s.renewBundle(ctx, domains, domain.BundleGroup); err != nil {
				failures = errors.Join(failures, err)
			} else {
				renewed += count
			}
			continue
		}

		var expires time.Time
		var err error

		if domain.VerificationMethod == DomainMethodDNSAutomation || domain.VerificationMethod == DomainMethodDNSDelegation {
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

func (s *DomainService) renewBundle(ctx context.Context, domains []models.Domain, group string) (int, error) {
	members := make([]models.Domain, 0, 2)
	for i := range domains {
		if domains[i].BundleGroup == group {
			members = append(members, domains[i])
		}
	}
	if len(members) == 0 {
		return 0, nil
	}
	hostnames := make([]string, 0, len(members))
	for _, member := range members {
		hostnames = append(hostnames, member.Hostname)
	}
	expires, err := s.issueAutomation(ctx, &members[0], hostnames)
	if err != nil {
		return 0, fmt.Errorf("renew bundle %s: %w", group, err)
	}
	for i := range members {
		members[i].CertificateExpiresAt = &expires
		members[i].CertificateStatus = "ready"
		if err := s.domains.Update(ctx, &members[i]); err != nil {
			return 0, fmt.Errorf("record %s renewal: %w", members[i].Hostname, err)
		}
	}
	return len(members), nil
}

func (s *DomainService) issueAutomation(ctx context.Context, domain *models.Domain, hostnames []string) (time.Time, error) {
	if s.dnsIssuer == nil {
		return time.Time{}, fmt.Errorf("automated issuance is not configured")
	}
	if domain.DelegationTarget != "" {
		if s.challenge == nil || s.challengeZone == "" {
			return time.Time{}, fmt.Errorf("delegated issuance is not configured")
		}
		var expires time.Time
		var err error
		if len(hostnames) > 1 {
			_, _, expires, err = s.dnsIssuer.IssueBundleDelegatedDNS01(ctx, hostnames, domain.DelegationTarget, s.challenge)
		} else {
			_, _, expires, err = s.dnsIssuer.IssueDelegatedDNS01(ctx, hostnames[0], domain.DelegationTarget, s.challenge)
		}
		if err != nil {
			return time.Time{}, fmt.Errorf("issue delegated certificate: %w", err)
		}
		return expires, nil
	}
	if s.dnsCredentials == nil {
		return time.Time{}, fmt.Errorf("automated issuance is not configured")
	}
	solver, err := s.dnsSolverForDomain(ctx, domain)
	if err != nil {
		return time.Time{}, err
	}
	var expires time.Time
	if len(hostnames) > 1 {
		_, _, expires, err = s.dnsIssuer.IssueBundleDNS01(ctx, hostnames, solver)
	} else {
		_, _, expires, err = s.dnsIssuer.IssueDNS01(ctx, hostnames[0], solver)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("issue automated certificate: %w", err)
	}
	s.dnsCredentials.TouchUsed(ctx, domain.OrganizationID)
	return expires, nil
}

func (s *DomainService) dnsSolverForDomain(ctx context.Context, domain *models.Domain) (dns.ChallengeSolver, error) {
	if s.dnsCredentials == nil {
		return nil, fmt.Errorf("dns credentials are not configured")
	}

	factory := s.dnsSolver

	if factory == nil {
		factory = SolverFor
	}

	if domain.DNSProvider != "" {
		token, err := s.dnsCredentials.Open(ctx, domain.OrganizationID, domain.DNSProvider)
		if err != nil {
			return nil, err
		}
		return factory(domain.DNSProvider, token)
	}

	credentials, err := s.dnsCredentials.List(ctx, domain.OrganizationID)
	if err != nil {
		return nil, err
	}
	if len(credentials) != 1 {
		return nil, fmt.Errorf("connect exactly one DNS provider to use automated verification")
	}
	token, err := s.dnsCredentials.Open(ctx, domain.OrganizationID, credentials[0].Provider)
	if err != nil {
		return nil, err
	}
	return factory(credentials[0].Provider, token)
}

func (s *DomainService) renewAutomated(ctx context.Context, domain *models.Domain) (time.Time, error) {
	expires, err := s.issueAutomation(ctx, domain, []string{domain.Hostname})
	if err != nil {
		return time.Time{}, fmt.Errorf("issue automated renewal: %w", err)
	}
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

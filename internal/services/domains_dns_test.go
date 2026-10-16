package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/infra/dns"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type stubDNSProtector struct{}

func (stubDNSProtector) Seal(value string) (string, error) { return "sealed:" + value, nil }
func (stubDNSProtector) Open(value string) (string, error) {
	return strings.TrimPrefix(value, "sealed:"), nil
}

type stubChallengeSolver struct {
	records map[string]string
}

func (s *stubChallengeSolver) Present(_ context.Context, fqdn, value string) error {
	if s.records == nil {
		s.records = map[string]string{}
	}
	s.records[fqdn] = value
	return nil
}

func (s *stubChallengeSolver) Cleanup(_ context.Context, fqdn string) error {
	delete(s.records, fqdn)
	return nil
}

func (s *stubChallengeSolver) Name() string { return "stub" }

type stubDNS01Issuer struct {
	hostnames []string
	expires   time.Time
	err       error
}

func (s *stubDNS01Issuer) issue(hostname string, solver dns.ChallengeSolver, ctx context.Context) ([]byte, []byte, time.Time, error) {
	if s.err != nil {
		return nil, nil, time.Time{}, s.err
	}
	s.hostnames = append(s.hostnames, hostname)
	if err := solver.Present(ctx, "_acme-challenge."+hostname, "digest"); err != nil {
		return nil, nil, time.Time{}, err
	}
	if err := solver.Cleanup(ctx, "_acme-challenge."+hostname); err != nil {
		return nil, nil, time.Time{}, err
	}
	return []byte("cert"), []byte("key"), s.expires, nil
}

func (s *stubDNS01Issuer) IssueDNS01(_ context.Context, hostname string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	return s.issue(hostname, solver, context.Background())
}

func (s *stubDNS01Issuer) IssueBundleDNS01(_ context.Context, hostnames []string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	for _, hostname := range hostnames {
		if _, _, _, err := s.issue(hostname, solver, context.Background()); err != nil {
			return nil, nil, time.Time{}, err
		}
	}
	return []byte("cert"), []byte("key"), s.expires, nil
}

func (s *stubDNS01Issuer) IssueDelegatedDNS01(_ context.Context, hostname string, _ string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	return s.issue(hostname, solver, context.Background())
}

func (s *stubDNS01Issuer) IssueBundleDelegatedDNS01(_ context.Context, hostnames []string, _ string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	for _, hostname := range hostnames {
		if _, _, _, err := s.issue(hostname, solver, context.Background()); err != nil {
			return nil, nil, time.Time{}, err
		}
	}
	return []byte("cert"), []byte("key"), s.expires, nil
}

func setupDomainServiceForDNS(t *testing.T) (*DomainService, *stubChallengeSolver, *stubDNS01Issuer) {
	t.Helper()
	db := reviewTestDatabase(t, &models.Domain{}, &models.DNSProviderCredential{})
	domainRepo, err := repositories.NewDomainRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	dnsRepo, err := repositories.NewDNSProviderRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewDomainService(domainRepo)
	if err != nil {
		t.Fatal(err)
	}
	credentials, err := NewDNSCredentialService(dnsRepo)
	if err != nil {
		t.Fatal(err)
	}
	credentials.SetProtector(stubDNSProtector{})
	svc.SetDNSCredentials(credentials)
	issuer := &stubDNS01Issuer{expires: time.Now().Add(90 * 24 * time.Hour)}
	svc.SetDNS01Issuer(issuer)
	solver := &stubChallengeSolver{}
	svc.dnsSolver = func(string, string) (dns.ChallengeSolver, error) { return solver, nil }
	return svc, solver, issuer
}

func TestDNSCredentialStoreRotateAndDelete(t *testing.T) {
	db := reviewTestDatabase(t, &models.DNSProviderCredential{})
	repo, err := repositories.NewDNSProviderRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewDNSCredentialService(repo)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Store(context.Background(), "org", "cloudflare", "token-a", "primary"); err == nil {
		t.Fatal("credential stored without a protector")
	}
	svc.SetProtector(stubDNSProtector{})
	if _, err := svc.Store(context.Background(), "org", "route53", "token", ""); err == nil {
		t.Fatal("unsupported provider was accepted")
	}
	first, err := svc.Store(context.Background(), "org", "cloudflare", "token-a", "primary")
	if err != nil {
		t.Fatal(err)
	}
	if first.LastFour != "en-a" || first.Ciphertext == "token-a" {
		t.Fatalf("credential was not encrypted with a suffix hint: %+v", first)
	}
	second, err := svc.Store(context.Background(), "org", "cloudflare", "token-b", "primary")
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || second.LastFour != "en-b" {
		t.Fatal("credential rotation did not reuse the provider row")
	}
	listed, err := svc.List(context.Background(), "org")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one credential, got %d", len(listed))
	}
	opened, err := svc.Open(context.Background(), "org", "cloudflare")
	if err != nil || opened != "token-b" {
		t.Fatalf("stored token did not round-trip: %q %v", opened, err)
	}
	if err := svc.Delete(context.Background(), "other-org", first.ID); err != repositories.ErrNotFound {
		t.Fatalf("cross-organization delete returned %v", err)
	}
	if err := svc.Delete(context.Background(), "org", first.ID); err != nil {
		t.Fatal(err)
	}
}

func TestWildcardDomainRequiresAutomation(t *testing.T) {
	svc, _, _ := setupDomainServiceForDNS(t)
	if _, _, err := svc.Create(context.Background(), "org", "*.example.com", "dns", nil); err == nil {
		t.Fatal("manual wildcard registration was accepted")
	}
	if _, _, err := svc.Create(context.Background(), "org", "*.example.com", "http", nil); err == nil {
		t.Fatal("manual wildcard registration was accepted")
	}
	if _, _, err := svc.Create(context.Background(), "org", "api.example.com", "carrier-pigeon", nil); err == nil {
		t.Fatal("unknown verification method was accepted")
	}
	if _, _, err := svc.Create(context.Background(), "org", "*.example.com", "dns-automation", nil); err != nil {
		t.Fatalf("automated wildcard registration rejected: %v", err)
	}
}

func TestAutomatedVerificationIssuesWildcard(t *testing.T) {
	svc, solver, issuer := setupDomainServiceForDNS(t)
	if _, err := svc.StoreDNSCredential(context.Background(), "org", "cloudflare", "cf-token", "primary"); err != nil {
		t.Fatal(err)
	}
	_, domain, err := svc.Create(context.Background(), "org", "*.example.com", "dns-automation", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Verify(context.Background(), domain.ID, ""); err != nil {
		t.Fatalf("automated verification failed: %v", err)
	}
	verified, err := svc.Find(context.Background(), domain.ID)
	if err != nil {
		t.Fatal(err)
	}
	if verified.Status != models.DomainStatusActive || verified.CertificateStatus != "ready" || verified.CertificateExpiresAt == nil || verified.VerifiedAt == nil {
		t.Fatalf("automated verification did not activate the domain: %+v", verified)
	}
	if len(issuer.hostnames) != 1 || issuer.hostnames[0] != "*.example.com" {
		t.Fatalf("issuer received %v", issuer.hostnames)
	}
	if len(solver.records) != 0 {
		t.Fatalf("challenge records were not cleaned up: %v", solver.records)
	}
}

func TestRenewDueCertificatesRenewsExpiring(t *testing.T) {
	svc, _, issuer := setupDomainServiceForDNS(t)
	if _, err := svc.StoreDNSCredential(context.Background(), "org", "cloudflare", "cf-token", "primary"); err != nil {
		t.Fatal(err)
	}
	_, domain, err := svc.Create(context.Background(), "org", "*.example.com", "dns-automation", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Verify(context.Background(), domain.ID, ""); err != nil {
		t.Fatal(err)
	}
	renewed, err := svc.RenewDueCertificates(context.Background(), 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if renewed != 0 {
		t.Fatalf("fresh certificate was renewed: %d", renewed)
	}
	soon := time.Now().Add(24 * time.Hour)
	stored, err := svc.Find(context.Background(), domain.ID)
	if err != nil {
		t.Fatal(err)
	}
	stored.CertificateExpiresAt = &soon
	if err := svc.domains.Update(context.Background(), &stored); err != nil {
		t.Fatal(err)
	}
	renewed, err = svc.RenewDueCertificates(context.Background(), 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if renewed != 1 || len(issuer.hostnames) != 2 {
		t.Fatalf("expiring certificate was not renewed: %d %v", renewed, issuer.hostnames)
	}
}

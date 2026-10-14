package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func writeTestCertificate(t *testing.T, dir, hostname string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: hostname},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(90 * 24 * time.Hour),
		DNSNames:     []string{hostname},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	certFile, keyFile := certificates.HostCertificatePaths(dir, hostname)
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
}

func setupServeCertificateService(t *testing.T) (*DomainService, string) {
	t.Helper()
	db := reviewTestDatabase(t, &models.Domain{})
	repo, err := repositories.NewDomainRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewDomainService(repo)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	service.SetCertificateCacheDir(dir)
	return service, dir
}

func seedActiveDomain(t *testing.T, service *DomainService, ctx context.Context, org, hostname string) models.Domain {
	t.Helper()
	_, domain, err := service.Create(ctx, org, hostname, DomainMethodHTTP, nil)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(90 * 24 * time.Hour)
	domain.Status = models.DomainStatusActive
	domain.CertificateStatus = "ready"
	domain.CertificateExpiresAt = &expires
	now := time.Now()
	domain.VerifiedAt = &now
	if err := service.domains.Update(ctx, &domain); err != nil {
		t.Fatal(err)
	}
	return domain
}

func TestServeCertificateExactAndWildcard(t *testing.T) {
	service, _ := setupServeCertificateService(t)
	ctx := context.Background()
	seedActiveDomain(t, service, ctx, "org-1", "app.example.com")
	expires := time.Now().Add(90 * 24 * time.Hour)
	now := time.Now()
	wildcard := models.Domain{OrganizationID: "org-1", Hostname: "*.example.com", Status: models.DomainStatusActive, VerificationMethod: DomainMethodDNSAutomation, CertificateStatus: "ready", CertificateExpiresAt: &expires, VerifiedAt: &now}
	if err := service.domains.Create(ctx, &wildcard); err != nil {
		t.Fatal(err)
	}
	writeTestCertificate(t, service.certCacheDir, "app.example.com")
	writeTestCertificate(t, service.certCacheDir, "*.example.com")

	exact, err := service.ServeCertificate(ctx, "org-1", "app.example.com")
	if err != nil {
		t.Fatalf("serve exact certificate: %v", err)
	}
	if exact.Hostname != "app.example.com" || exact.WildcardFor != "" {
		t.Fatalf("exact certificate = %+v", exact.Hostname)
	}
	if _, err := tls.X509KeyPair(exact.ChainPEM, exact.KeyPEM); err != nil {
		t.Fatalf("exact pair does not parse: %v", err)
	}

	wild, err := service.ServeCertificate(ctx, "org-1", "api.example.com")
	if err != nil {
		t.Fatalf("serve wildcard certificate: %v", err)
	}
	if wild.Hostname != "*.example.com" || wild.WildcardFor != "api.example.com" {
		t.Fatalf("wildcard certificate = %+v", wild.Hostname)
	}
}

func TestServeCertificateEnforcesOrgScopeAndState(t *testing.T) {
	service, _ := setupServeCertificateService(t)
	ctx := context.Background()
	seedActiveDomain(t, service, ctx, "org-1", "app.example.com")
	writeTestCertificate(t, service.certCacheDir, "app.example.com")

	if _, err := service.ServeCertificate(ctx, "org-2", "app.example.com"); err == nil {
		t.Fatal("foreign organization received certificate material")
	}
	if _, err := service.ServeCertificate(ctx, "", "app.example.com"); err == nil {
		t.Fatal("missing organization received certificate material")
	}
	if _, err := service.ServeCertificate(ctx, "org-1", "unknown.example.com"); err == nil {
		t.Fatal("unknown hostname received certificate material")
	}
	if _, err := service.ServeCertificate(ctx, "org-1", "*.example.com"); err == nil {
		t.Fatal("wildcard pattern received certificate material")
	}
	if _, err := service.ServeCertificate(ctx, "org-1", "APP.EXAMPLE.COM."); err != nil {
		t.Fatalf("hostname normalization failed: %v", err)
	}
}

func TestServeCertificateRejectsUnreadyAndRevoked(t *testing.T) {
	service, _ := setupServeCertificateService(t)
	ctx := context.Background()
	if _, _, err := service.Create(ctx, "org-1", "pending.example.com", DomainMethodHTTP, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ServeCertificate(ctx, "org-1", "pending.example.com"); err == nil {
		t.Fatal("pending domain served a certificate")
	}

	domain := seedActiveDomain(t, service, ctx, "org-1", "gone.example.com")
	writeTestCertificate(t, service.certCacheDir, "gone.example.com")
	certFile, keyFile := certificates.HostCertificatePaths(service.certCacheDir, "gone.example.com")
	if err := service.RevokeDomainCertificate(ctx, domain.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ServeCertificate(ctx, "org-1", "gone.example.com"); err == nil {
		t.Fatal("revoked domain served a certificate")
	}
	if _, err := os.Stat(certFile); !os.IsNotExist(err) {
		t.Fatal("revoked certificate file was not removed")
	}
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		t.Fatal("revoked key file was not removed")
	}
	if _, err := os.Stat(filepath.Join(service.certCacheDir, "unrelated.crt")); !os.IsNotExist(err) {
		t.Fatal("unexpected file in certificate cache")
	}
}

func TestServeCertificateRejectsMissingFiles(t *testing.T) {
	service, _ := setupServeCertificateService(t)
	ctx := context.Background()
	seedActiveDomain(t, service, ctx, "org-1", "nofiles.example.com")
	if _, err := service.ServeCertificate(ctx, "org-1", "nofiles.example.com"); err == nil {
		t.Fatal("domain without certificate files served material")
	}
}

func TestDeleteDomainRemovesCertificateFiles(t *testing.T) {
	service, _ := setupServeCertificateService(t)
	ctx := context.Background()
	domain := seedActiveDomain(t, service, ctx, "org-1", "bye.example.com")
	writeTestCertificate(t, service.certCacheDir, "bye.example.com")
	if err := service.Delete(ctx, domain.ID); err != nil {
		t.Fatal(err)
	}
	certFile, keyFile := certificates.HostCertificatePaths(service.certCacheDir, "bye.example.com")
	if _, err := os.Stat(certFile); !os.IsNotExist(err) {
		t.Fatal("deleted domain certificate file was not removed")
	}
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		t.Fatal("deleted domain key file was not removed")
	}
}

func TestHostCertificatePaths(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := certificates.HostCertificatePaths(dir, "App.Example.COM.")
	if certFile != filepath.Join(dir, "app_example_com.crt") || keyFile != filepath.Join(dir, "app_example_com.key") {
		t.Fatalf("paths = %s %s", certFile, keyFile)
	}
	wildCert, wildKey := certificates.HostCertificatePaths(dir, "*.example.com")
	if wildCert != filepath.Join(dir, "wildcard_example_com.crt") || wildKey != filepath.Join(dir, "wildcard_example_com.key") {
		t.Fatalf("wildcard paths = %s %s", wildCert, wildKey)
	}
}

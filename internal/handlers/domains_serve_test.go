package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func newServeCertificateStack(t *testing.T) (*DomainHandler, *services.DomainService, *repositories.GormDomainRepository, *gorm.DB, string) {
	t.Helper()
	connection, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	db, err := gorm.Open(sqlite.Dialector{Conn: connection}, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Domain{}, &models.Tunnel{}); err != nil {
		t.Fatal(err)
	}
	domainRepo, err := repositories.NewDomainRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewDomainService(domainRepo)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	service.SetCertificateCacheDir(dir)
	handler, err := NewDomainHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	return handler, service, domainRepo, db, dir
}

func writeServeCertificate(t *testing.T, dir, hostname string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: hostname}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour), DNSNames: []string{hostname}}
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

func TestServeCertificateInternalEndpoint(t *testing.T) {
	handler, service, domainRepo, _, dir := newServeCertificateStack(t)
	ctx := context.Background()
	_, domain, err := service.Create(ctx, "org-1", "app.example.com", services.DomainMethodHTTP, nil)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(24 * time.Hour)
	domain.Status = models.DomainStatusActive
	domain.CertificateStatus = "ready"
	domain.CertificateExpiresAt = &expires
	if err := domainRepo.Update(ctx, &domain); err != nil {
		t.Fatal(err)
	}
	writeServeCertificate(t, dir, "app.example.com")

	app := fiber.New()
	app.Get("/internal/domains/certificate", handler.ServeCertificate)

	serve := func(query string) (int, map[string]any) {
		request := httptest.NewRequest("GET", "/internal/domains/certificate"+query, nil)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var result map[string]any
		if response.StatusCode == 200 {
			if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
		}
		return response.StatusCode, result
	}

	if status, result := serve("?hostname=app.example.com"); status != 200 || result["hostname"] != "app.example.com" {
		t.Fatalf("serve returned %d: %v", status, result)
	} else {
		if _, ok := result["certificate"].(string); !ok {
			t.Fatalf("response missing chain material: %v", result)
		}
		if _, ok := result["key"].(string); !ok {
			t.Fatalf("response missing key material: %v", result)
		}
		if _, ok := result["expiresAt"].(string); !ok {
			t.Fatalf("response missing expiry: %v", result)
		}
	}
	if status, _ := serve("?hostname=other.example.com"); status != 404 {
		t.Fatalf("unknown hostname returned %d", status)
	}
	if status, _ := serve(""); status != 400 {
		t.Fatalf("missing identity returned %d", status)
	}
}

func TestServeCertificateTunnelScope(t *testing.T) {
	handler, service, domainRepo, db, dir := newServeCertificateStack(t)
	ctx := context.Background()
	tunnelRepo, err := repositories.NewTunnelRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tunnels, err := services.NewTunnelService(tunnelRepo)
	if err != nil {
		t.Fatal(err)
	}
	handler.SetTunnels(tunnels)
	tunnel, err := tunnels.Create(ctx, "org-1", "web", models.TunnelProtocolHTTP, "127.0.0.1", 3000, "app.example.com", "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, domain, err := service.Create(ctx, "org-1", "app.example.com", services.DomainMethodHTTP, &tunnel.ID)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(24 * time.Hour)
	domain.Status = models.DomainStatusActive
	domain.CertificateStatus = "ready"
	domain.CertificateExpiresAt = &expires
	if err := domainRepo.Update(ctx, &domain); err != nil {
		t.Fatal(err)
	}
	writeServeCertificate(t, dir, "app.example.com")

	app := fiber.New()
	app.Get("/internal/domains/certificate", handler.ServeCertificate)
	serve := func(query string) int {
		request := httptest.NewRequest("GET", "/internal/domains/certificate"+query, nil)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		return response.StatusCode
	}

	if status := serve("?tunnelId=" + tunnel.ID); status != 200 {
		t.Fatalf("tunnel-scoped serve returned %d", status)
	}
	if status := serve("?tunnelId=" + tunnel.ID + "&hostname=app.example.com"); status != 200 {
		t.Fatalf("tunnel plus hostname serve returned %d", status)
	}
	if status := serve("?tunnelId=" + tunnel.ID + "&hostname=evil.example.com"); status != 403 {
		t.Fatalf("hostname mismatch returned %d", status)
	}
	if status := serve("?tunnelId=missing"); status != 404 {
		t.Fatalf("unknown tunnel returned %d", status)
	}
}

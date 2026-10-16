package relay

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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func serveFetcherCertificate(t *testing.T, hostname string) (string, string) {
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
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certPEM, keyPEM
}

func TestInternalCertificateFetcherCachesAndScopes(t *testing.T) {
	certPEM, keyPEM := serveFetcherCertificate(t, "app.example.com")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/internal/domains/certificate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-Internal-Secret") != "internal-secret" {
			t.Errorf("missing internal secret")
		}
		if hostname := r.URL.Query().Get("hostname"); hostname != "app.example.com" {
			t.Errorf("hostname = %s", hostname)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hostname":"app.example.com","certificate":` + pemJSON(certPEM) + `,"key":` + pemJSON(keyPEM) + `,"expiresAt":"2027-01-01T00:00:00Z","wildcardFor":""}`))
	}))
	defer server.Close()
	fetcher, err := NewInternalCertificateFetcher(server.URL, "internal-secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	first, err := fetcher.CertificateFor(context.Background(), "App.Example.COM.")
	if err != nil {
		t.Fatal(err)
	}
	second, err := fetcher.CertificateForHostname("app.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("second lookup did not hit the certificate cache")
	}
	if calls.Load() != 1 {
		t.Fatalf("api calls = %d", calls.Load())
	}
	if _, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM)); err != nil {
		t.Fatal(err)
	}
}

func pemJSON(pem string) string {
	escaped := strings.ReplaceAll(pem, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	return "\"" + escaped + "\""
}

func TestInternalCertificateFetcherRejectsErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	fetcher, err := NewInternalCertificateFetcher(server.URL, "internal-secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fetcher.CertificateFor(context.Background(), "missing.example.com"); err == nil {
		t.Fatal("missing certificate was accepted")
	}
	if _, err := NewInternalCertificateFetcher("", "", nil); err == nil {
		t.Fatal("fetcher accepted empty configuration")
	}
}

func TestRelayServesCustomDomainOverSNI(t *testing.T) {
	certPEM, keyPEM := serveFetcherCertificate(t, "app.example.com")
	var calls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hostname":"app.example.com","certificate":` + pemJSON(certPEM) + `,"key":` + pemJSON(keyPEM) + `,"expiresAt":"2027-01-01T00:00:00Z","wildcardFor":""}`))
	}))
	defer api.Close()
	fetcher, err := NewInternalCertificateFetcher(api.URL, "internal-secret", nil)
	if err != nil {
		t.Fatal(err)
	}
	hello := &tls.ClientHelloInfo{ServerName: "app.example.com"}
	certificate, err := fetcher.CertificateForHostname(hello.ServerName)
	if err != nil {
		t.Fatal(err)
	}
	if len(certificate.Certificate) == 0 {
		t.Fatal("fetched certificate is empty")
	}
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := leaf.VerifyHostname("app.example.com"); err != nil {
		t.Fatalf("leaf does not cover SNI hostname: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("api calls = %d", calls.Load())
	}
}

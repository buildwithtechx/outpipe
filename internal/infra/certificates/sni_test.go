package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeDispatcherCertificate(t *testing.T, dir, name, hostname string) (string, string) {
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
	certFile := filepath.Join(dir, name+".crt")
	keyFile := filepath.Join(dir, name+".key")
	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}

type stubCertificateProvider struct {
	cert  *tls.Certificate
	hosts map[string]bool
	err   error
	calls []string
}

func (s *stubCertificateProvider) CertificateForHostname(hostname string) (*tls.Certificate, error) {
	s.calls = append(s.calls, hostname)
	if s.err != nil {
		return nil, s.err
	}
	if s.hosts != nil && !s.hosts[hostname] {
		return nil, errors.New("no certificate for hostname")
	}
	return s.cert, nil
}

func TestSNIDispatcherServesPerHostAndFallsBack(t *testing.T) {
	dir := t.TempDir()
	fallbackCert, fallbackKey := writeDispatcherCertificate(t, dir, "fallback", "relay.example.com")
	fallback, err := NewRotatingCertificate(fallbackCert, fallbackKey)
	if err != nil {
		t.Fatal(err)
	}
	customCert, customKey := writeDispatcherCertificate(t, dir, "custom", "app.example.com")
	custom, err := tls.LoadX509KeyPair(customCert, customKey)
	if err != nil {
		t.Fatal(err)
	}
	provider := &stubCertificateProvider{cert: &custom, hosts: map[string]bool{"app.example.com": true}}
	dispatcher, err := NewSNIDispatcher(fallback, provider)
	if err != nil {
		t.Fatal(err)
	}

	served, err := dispatcher.GetCertificate(&tls.ClientHelloInfo{ServerName: "app.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if len(served.Certificate) == 0 || len(custom.Certificate) == 0 || string(served.Certificate[0]) != string(custom.Certificate[0]) {
		t.Fatal("SNI hostname did not receive its certificate")
	}

	served, err = dispatcher.GetCertificate(&tls.ClientHelloInfo{ServerName: "unknown.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	servedFallback, err := fallback.GetCertificate(&tls.ClientHelloInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if string(served.Certificate[0]) != string(servedFallback.Certificate[0]) {
		t.Fatal("unknown hostname did not fall back to the mounted certificate")
	}

	if _, err := dispatcher.GetCertificate(&tls.ClientHelloInfo{}); err != nil {
		t.Fatalf("empty SNI failed: %v", err)
	}
	if len(provider.calls) != 2 || provider.calls[0] != "app.example.com" || provider.calls[1] != "unknown.example.com" {
		t.Fatalf("provider calls = %v", provider.calls)
	}
}

func TestSNIDispatcherFallsBackWhenProviderFails(t *testing.T) {
	dir := t.TempDir()
	fallbackCert, fallbackKey := writeDispatcherCertificate(t, dir, "fallback", "relay.example.com")
	fallback, err := NewRotatingCertificate(fallbackCert, fallbackKey)
	if err != nil {
		t.Fatal(err)
	}
	dispatcher, err := NewSNIDispatcher(fallback, &stubCertificateProvider{err: errors.New("api unavailable")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dispatcher.GetCertificate(&tls.ClientHelloInfo{ServerName: "app.example.com"}); err != nil {
		t.Fatalf("provider failure broke the fallback: %v", err)
	}
	if _, err := NewSNIDispatcher(nil, nil); err == nil {
		t.Fatal("dispatcher accepted a nil fallback")
	}
}

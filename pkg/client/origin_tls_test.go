package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

func TestOriginTLSForwardsOnlyWithTrustedCAAndMatchingName(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate origin key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), DNSNames: []string{"localhost"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create origin certificate: %v", err)
	}
	privateKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("encode origin key: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	certificate, err := tls.X509KeyPair(certPEM, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKey}))
	if err != nil {
		t.Fatalf("load origin certificate: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || r.TLS.ServerName != "localhost" {
			t.Error("expected verified localhost TLS origin")
		}
		if _, err := w.Write([]byte("local TLS response")); err != nil {
			t.Errorf("write origin response: %v", err)
		}
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12}
	server.StartTLS()
	defer server.Close()
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(caFile, certPEM, 0600); err != nil {
		t.Fatalf("write origin CA: %v", err)
	}
	for _, test := range []struct {
		name    string
		config  OriginTLSConfig
		success bool
	}{
		{"trusted localhost", OriginTLSConfig{CAFile: caFile, ServerName: "localhost"}, true},
		{"untrusted CA", OriginTLSConfig{ServerName: "localhost"}, false},
		{"wrong hostname", OriginTLSConfig{CAFile: caFile, ServerName: "wrong.example"}, false},
		{"IP mismatch", OriginTLSConfig{CAFile: caFile}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			origin, err := NewOriginHTTPClient(test.config)
			if err != nil {
				t.Fatalf("configure origin: %v", err)
			}
			defer origin.CloseIdleConnections()
			connection := &RelayConnection{}
			response := connection.forwardHTTP(context.Background(), server.URL, protocol.HTTPRequest{Method: http.MethodGet, Path: "/"}, origin)
			if test.success {
				if response.Error != "" || response.StatusCode != http.StatusOK || response.Body != base64.StdEncoding.EncodeToString([]byte("local TLS response")) {
					t.Fatalf("expected local TLS response, got %+v", response)
				}
			} else if response.Error == "" {
				t.Fatal("untrusted origin was accepted")
			}
		})
	}
}

func TestOriginTLSRejectsInvalidConfiguration(t *testing.T) {
	invalidCA := filepath.Join(t.TempDir(), "invalid.pem")
	if err := os.WriteFile(invalidCA, []byte("not a certificate"), 0600); err != nil {
		t.Fatalf("write invalid CA: %v", err)
	}
	for _, config := range []OriginTLSConfig{
		{CAFile: invalidCA}, {CAFile: filepath.Join(t.TempDir(), "missing.pem")}, {ServerName: "https://localhost"},
	} {
		if _, err := NewOriginHTTPClient(config); err == nil {
			t.Fatalf("accepted invalid origin config %+v", config)
		}
	}
}

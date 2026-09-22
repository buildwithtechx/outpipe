package certificates

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewACMEIssuerRequiresConfiguration(t *testing.T) {

	if _, err := NewACMEIssuer(ACMEConfig{Email: "ops@example.com"}); err == nil {
		t.Fatal("expected an error when the acme cache directory is missing")
	}

	if _, err := NewACMEIssuer(ACMEConfig{CacheDir: t.TempDir()}); err == nil {
		t.Fatal("expected an error when the acme email is missing")
	}

	issuer, err := NewACMEIssuer(ACMEConfig{Email: "ops@example.com", CacheDir: t.TempDir()})

	if err != nil {
		t.Fatalf("expected a valid issuer: %v", err)
	}

	if issuer == nil || issuer.manager == nil {
		t.Fatal("expected a configured issuer")
	}
}

func TestACMEIssuerHostPolicy(t *testing.T) {
	issuer, err := NewACMEIssuer(ACMEConfig{Email: "ops@example.com", CacheDir: t.TempDir(), AllowedHost: func(host string) bool {
		return strings.Contains(host, ".") && !strings.Contains(host, "..")
	}})

	if err != nil {
		t.Fatal(err)
	}

	policy := issuer.manager.HostPolicy

	if err := policy(nil, "evil..example.com"); err == nil {
		t.Fatal("expected hosts with empty labels to be rejected by host policy")
	}

	if err := policy(nil, "localhost"); err == nil {
		t.Fatal("expected dot-less hosts to be rejected by host policy")
	}

	if err := policy(nil, "myapp.outpipe.app"); err != nil {
		t.Fatalf("expected a subdomain to be allowed: %v", err)
	}

	if err := policy(nil, "custom.example.com"); err != nil {
		t.Fatalf("expected a custom domain to be allowed: %v", err)
	}
}

func TestTLSListenerServesCertificate(t *testing.T) {
	directory := t.TempDir()
	certFile := filepath.Join(directory, "cert.pem")
	keyFile := filepath.Join(directory, "key.pem")

	if err := writeCertificate(certFile, keyFile, "tunnel.example.com"); err != nil {
		t.Fatal(err)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("tls-ok"))
	}))
	defer upstream.Close()

	listener, err := NewTLSListener("127.0.0.1:0", certFile, keyFile)

	if err != nil {
		t.Fatalf("create tls listener: %v", err)
	}

	defer listener.Close()
	go func() {
		_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, requestErr := http.Get(upstream.URL)
			_ = requestErr

			if requestErr != nil {
				http.Error(w, "upstream unavailable", http.StatusBadGateway)
				return
			}

			defer response.Body.Close()
			_, _ = w.Write([]byte("tls-ok"))
		}))
	}()

	pool := x509.NewCertPool()
	certificate, err := os.ReadFile(certFile)

	if err != nil {
		t.Fatal(err)
	}

	if !pool.AppendCertsFromPEM(certificate) {
		t.Fatal("failed to load test certificate into the pool")
	}

	connection, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{RootCAs: pool, ServerName: "tunnel.example.com", MinVersion: tls.VersionTLS12})

	if err != nil {
		t.Fatalf("tls handshake failed: %v", err)
	}

	defer connection.Close()
	state := connection.ConnectionState()

	if state.Version < tls.VersionTLS12 {
		t.Fatalf("unexpected tls version %d", state.Version)
	}

	_ = connection.SetDeadline(time.Now().Add(2 * time.Second))
	_, _ = connection.Write([]byte("GET / HTTP/1.1\r\nHost: tunnel.example.com\r\nConnection: close\r\n\r\n"))
	responseBytes := make([]byte, 0, 64)
	buffer := make([]byte, 64)

	for {
		count, readErr := connection.Read(buffer)

		if count > 0 {
			responseBytes = append(responseBytes, buffer[:count]...)
		}

		if readErr != nil {
			break
		}
	}

	if !bytes.Contains(responseBytes, []byte("tls-ok")) {
		t.Fatalf("unexpected response %q", responseBytes)
	}
}

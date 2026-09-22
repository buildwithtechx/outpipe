package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotatingCertificateReloadsChangedFiles(t *testing.T) {
	directory := t.TempDir()
	certFile := filepath.Join(directory, "cert.pem")
	keyFile := filepath.Join(directory, "key.pem")

	if err := writeCertificate(certFile, keyFile, "first.example.com"); err != nil {
		t.Fatal(err)
	}

	rotating, err := NewRotatingCertificate(certFile, keyFile)

	if err != nil {
		t.Fatal(err)
	}

	first, err := rotating.GetCertificate(nil)

	if err != nil {
		t.Fatal(err)
	}

	firstName := first.Leaf

	if firstName == nil {
		firstName, err = x509.ParseCertificate(first.Certificate[0])

		if err != nil {
			t.Fatal(err)
		}
	}

	if firstName.Subject.CommonName != "first.example.com" {
		t.Fatalf("unexpected initial certificate: %s", firstName.Subject.CommonName)
	}

	if err := writeCertificate(certFile, keyFile, "second.example.com"); err != nil {
		t.Fatal(err)
	}

	second, err := rotating.GetCertificate(nil)

	if err != nil {
		t.Fatal(err)
	}

	parsed, err := x509.ParseCertificate(second.Certificate[0])

	if err != nil {
		t.Fatal(err)
	}

	if parsed.Subject.CommonName != "second.example.com" {
		t.Fatalf("certificate did not rotate: %s", parsed.Subject.CommonName)
	}
}

func writeCertificate(certFile, keyFile, commonName string) error {
	key, err := rsa.GenerateKey(rand.Reader, 1024)

	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))

	if err != nil {
		return err
	}

	certificate, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: commonName}, DNSNames: []string{commonName}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour)}, &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: commonName}, DNSNames: []string{commonName}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour)}, &key.PublicKey, key)

	if err != nil {
		return err
	}

	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}), 0o600); err != nil {
		return err
	}

	return os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0o600)
}

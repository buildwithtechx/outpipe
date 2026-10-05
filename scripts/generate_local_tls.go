package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := generateLocalTLS(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generateLocalTLS() error {
	directory := filepath.Join("data", "tls")
	certPath, keyPath := filepath.Join(directory, "localhost.crt"), filepath.Join(directory, "localhost.key")
	if _, err := os.Stat(certPath); err == nil {
		if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
			return fmt.Errorf("validate existing local TLS pair: %w", err)
		}
		fmt.Println("Existing local TLS pair preserved")
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect local certificate: %w", err)
	}
	if _, err := os.Stat(keyPath); err == nil {
		return fmt.Errorf("local TLS key exists without a certificate; refusing to replace it")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect local TLS key: %w", err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate local TLS key: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate certificate serial: %w", err)
	}
	certificate := x509.Certificate{
		SerialNumber: serial, Subject: pkix.Name{CommonName: "Outpipe local development"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().AddDate(1, 0, 0),
		DNSNames:    []string{"localhost", "outpipe.localhost", "*.outpipe.localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		KeyUsage:    x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &certificate, &certificate, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("create local TLS certificate: %w", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("encode local TLS key: %w", err)
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return fmt.Errorf("create local TLS directory: %w", err)
	}
	if err := saveNewFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		return fmt.Errorf("save local TLS key: %w", err)
	}
	if err := saveNewFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0644); err != nil {
		return fmt.Errorf("save local TLS certificate: %w", err)
	}
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return fmt.Errorf("validate generated TLS pair: %w", err)
	}
	fmt.Println("Generated local TLS pair in data/tls")
	return nil
}

func saveNewFile(path string, data []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("create new TLS file: %w", err)
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write TLS file: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close TLS file: %w", closeErr)
	}
	return nil
}

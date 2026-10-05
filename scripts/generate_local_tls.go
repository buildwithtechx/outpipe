package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
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
		if err := validateLocalTLSPair(certPath, keyPath); err != nil {
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
	if err := saveLocalTLSPair(certPath, keyPath, der, keyDER); err != nil {
		return err
	}
	if err := validateLocalTLSPair(certPath, keyPath); err != nil {
		return fmt.Errorf("validate generated TLS pair: %w", err)
	}
	fmt.Println("Generated local TLS pair in data/tls")
	return nil
}

func saveLocalTLSPair(certPath, keyPath string, der, keyDER []byte) error {
	keyInfo, err := saveNewFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0600)
	if err != nil {
		return fmt.Errorf("save local TLS key: %w", err)
	}
	if _, err := saveNewFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0644); err != nil {
		return fmt.Errorf("save local TLS certificate: %w", errors.Join(err, removeCreatedFile(keyPath, keyInfo)))
	}
	return nil
}

func saveNewFile(path string, data []byte, mode os.FileMode) (os.FileInfo, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return nil, fmt.Errorf("create new TLS file: %w", err)
	}
	info, statErr := file.Stat()
	if statErr != nil {
		return nil, fmt.Errorf("inspect new TLS file: %w", errors.Join(statErr, file.Close()))
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return nil, fmt.Errorf("save TLS file: %w", errors.Join(err, removeCreatedFile(path, info)))
	}
	return info, nil
}

func removeCreatedFile(path string, created os.FileInfo) error {
	current, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect incomplete TLS file: %w", err)
	}
	if !os.SameFile(current, created) {
		return fmt.Errorf("incomplete TLS file was replaced; refusing to remove it")
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove incomplete TLS file: %w", err)
	}
	return nil
}

func validateLocalTLSPair(certPath, keyPath string) error {
	pair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("load TLS pair: %w", err)
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return fmt.Errorf("parse TLS certificate: %w", err)
	}
	now := time.Now()
	if now.Before(leaf.NotBefore) || !now.Before(leaf.NotAfter) {
		return fmt.Errorf("local TLS certificate is expired or not yet valid; move the existing pair aside to regenerate")
	}
	for _, host := range []string{"localhost", "outpipe.localhost", "preview.outpipe.localhost", "127.0.0.1", "::1"} {
		if err := leaf.VerifyHostname(host); err != nil {
			return fmt.Errorf("local TLS certificate does not cover %s: %w", host, err)
		}
	}
	return nil
}

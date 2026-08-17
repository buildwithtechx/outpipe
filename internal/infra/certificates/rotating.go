package certificates

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sync"
)

type RotatingCertificate struct {
	certFile string
	keyFile  string
	mu       sync.RWMutex
	cert     *tls.Certificate
	certInfo os.FileInfo
	keyInfo  os.FileInfo
}

func NewRotatingCertificate(certFile, keyFile string) (*RotatingCertificate, error) {

	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("certificate and key files are required")
	}

	rotating := &RotatingCertificate{certFile: certFile, keyFile: keyFile}

	if err := rotating.reload(); err != nil {
		return nil, err
	}

	return rotating, nil
}

func (r *RotatingCertificate) Config() *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS12, GetCertificate: r.GetCertificate}
}

func (r *RotatingCertificate) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {

	if err := r.reloadIfChanged(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cert == nil {
		return nil, fmt.Errorf("certificate is not loaded")
	}

	return r.cert, nil
}

func NewTLSListener(address, certFile, keyFile string) (net.Listener, error) {
	rotating, err := NewRotatingCertificate(certFile, keyFile)

	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", address)

	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", address, err)
	}

	return tls.NewListener(listener, rotating.Config()), nil
}

func (r *RotatingCertificate) reloadIfChanged() error {
	certInfo, err := os.Stat(r.certFile)

	if err != nil {
		return fmt.Errorf("stat certificate file: %w", err)
	}

	keyInfo, err := os.Stat(r.keyFile)

	if err != nil {
		return fmt.Errorf("stat key file: %w", err)
	}

	r.mu.RLock()
	changed := r.certInfo == nil || r.keyInfo == nil || !sameFile(certInfo, r.certInfo) || !sameFile(keyInfo, r.keyInfo)
	r.mu.RUnlock()

	if !changed {
		return nil
	}

	return r.reload()
}

func (r *RotatingCertificate) reload() error {
	certificate, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)

	if err != nil {
		return fmt.Errorf("load TLS certificate: %w", err)
	}

	certInfo, err := os.Stat(r.certFile)

	if err != nil {
		return fmt.Errorf("stat certificate file: %w", err)
	}

	keyInfo, err := os.Stat(r.keyFile)

	if err != nil {
		return fmt.Errorf("stat key file: %w", err)
	}

	r.mu.Lock()
	r.cert = &certificate
	r.certInfo = certInfo
	r.keyInfo = keyInfo
	r.mu.Unlock()
	return nil
}

func sameFile(left, right os.FileInfo) bool {
	return left.Size() == right.Size() && left.ModTime().Equal(right.ModTime())
}

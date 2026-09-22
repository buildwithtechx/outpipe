package certificates

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type ACMEConfig struct {
	Email       string
	Directory   string
	CacheDir    string
	AllowedHost func(string) bool
}

type ACMEIssuer struct {
	manager *autocert.Manager
}

func NewACMEIssuer(cfg ACMEConfig) (*ACMEIssuer, error) {

	if strings.TrimSpace(cfg.Email) == "" || strings.TrimSpace(cfg.CacheDir) == "" {
		return nil, fmt.Errorf("acme email and cache directory are required")
	}

	if err := os.MkdirAll(cfg.CacheDir, 0o750); err != nil {
		return nil, fmt.Errorf("create acme cache: %w", err)
	}

	manager := &autocert.Manager{Prompt: autocert.AcceptTOS, Email: cfg.Email, Cache: autocert.DirCache(cfg.CacheDir)}

	if cfg.Directory != "" {
		manager.Client = &acme.Client{DirectoryURL: cfg.Directory}
	}

	manager.HostPolicy = func(_ context.Context, host string) error {

		if cfg.AllowedHost != nil && !cfg.AllowedHost(strings.ToLower(strings.TrimSuffix(host, "."))) {
			return fmt.Errorf("certificate host is not allowed")
		}

		return nil
	}
	return &ACMEIssuer{manager: manager}, nil
}

func (i *ACMEIssuer) Issue(ctx context.Context, hostname string) (time.Time, error) {

	if i == nil || i.manager == nil || hostname == "" {
		return time.Time{}, fmt.Errorf("acme issuer and hostname are required")
	}

	_ = ctx
	certificate, err := i.manager.GetCertificate(&tls.ClientHelloInfo{ServerName: hostname})

	if err != nil {
		return time.Time{}, fmt.Errorf("issue certificate: %w", err)
	}

	if len(certificate.Certificate) == 0 {
		return time.Time{}, fmt.Errorf("acme returned an empty certificate")
	}

	parsed, err := x509.ParseCertificate(certificate.Certificate[0])

	if err != nil {
		return time.Time{}, fmt.Errorf("parse issued certificate: %w", err)
	}

	return parsed.NotAfter, nil
}

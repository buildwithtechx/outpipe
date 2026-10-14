package certificates

import (
	"crypto/tls"
	"fmt"
	"strings"
)

type CertificateProvider interface {
	CertificateForHostname(hostname string) (*tls.Certificate, error)
}

type SNIDispatcher struct {
	fallback *RotatingCertificate
	provider CertificateProvider
}

func NewSNIDispatcher(fallback *RotatingCertificate, provider CertificateProvider) (*SNIDispatcher, error) {
	if fallback == nil {
		return nil, fmt.Errorf("fallback certificate is required")
	}
	return &SNIDispatcher{fallback: fallback, provider: provider}, nil
}

func (d *SNIDispatcher) Config() *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS12, GetCertificate: d.GetCertificate}
}

func (d *SNIDispatcher) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	serverName := ""
	if hello != nil {
		serverName = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hello.ServerName, ".")))
	}
	if d.provider != nil && serverName != "" {
		if certificate, err := d.provider.CertificateForHostname(serverName); err == nil && certificate != nil {
			return certificate, nil
		}
	}
	return d.fallback.GetCertificate(hello)
}

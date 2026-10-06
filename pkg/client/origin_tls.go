package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type OriginTLSConfig struct {
	CAFile     string `json:"caFile,omitempty"`
	ServerName string `json:"serverName,omitempty"`
}

func NewOriginHTTPClient(config OriginTLSConfig) (*http.Client, error) {
	if strings.ContainsAny(config.ServerName, " /\\\t\r\n") {
		return nil, fmt.Errorf("origin TLS server name must be a hostname or IP address")
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: config.ServerName}
	if config.CAFile != "" {
		certificates, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read origin CA file: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("load system CA certificates: %w", err)
		}
		if !roots.AppendCertsFromPEM(certificates) {
			return nil, fmt.Errorf("origin CA file contains no valid PEM certificates")
		}
		tlsConfig.RootCAs = roots
	}
	transport := &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: 10 * time.Second,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	}
	return &http.Client{Timeout: 90 * time.Second, Transport: transport}, nil
}

package certificates

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/acme"
	"outpipe.dev/outpipe/internal/infra/dns"
)

type DNS01Issuer struct {
	config ACMEConfig
}

func NewDNS01Issuer(config ACMEConfig) (*DNS01Issuer, error) {
	if strings.TrimSpace(config.Email) == "" || !strings.Contains(config.Email, "@") {
		return nil, fmt.Errorf("acme email is required")
	}
	if strings.TrimSpace(config.CacheDir) == "" {
		return nil, fmt.Errorf("acme cache directory is required")
	}
	if err := os.MkdirAll(config.CacheDir, 0o750); err != nil {
		return nil, fmt.Errorf("create acme cache: %w", err)
	}
	return &DNS01Issuer{config: config}, nil
}

func (i *DNS01Issuer) accountKey() (crypto.Signer, error) {
	path := filepath.Join(i.config.CacheDir, "acme_account.key")
	if raw, err := os.ReadFile(path); err == nil {
		block, _ := pem.Decode(raw)
		if block == nil {
			return nil, fmt.Errorf("acme account key is not valid PEM; delete %s to re-register", path)
		}
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("acme account key is invalid; delete %s to re-register: %w", path, err)
		}
		return key, nil
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate acme account key: %w", err)
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("encode acme account key: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("store acme account key: %w", err)
	}
	if err := pem.Encode(file, &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("write acme account key: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close acme account key: %w", err)
	}
	return key, nil
}

func (i *DNS01Issuer) acmeClient() (*acme.Client, error) {
	key, err := i.accountKey()
	if err != nil {
		return nil, err
	}
	directory := strings.TrimSpace(i.config.Directory)
	if directory == "" {
		directory = acme.LetsEncryptURL
	}
	return &acme.Client{Key: key, DirectoryURL: directory}, nil
}

func normalizeHostname(hostname string) (string, error) {
	hostname = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))
	if hostname == "" {
		return "", fmt.Errorf("hostname is required")
	}
	return hostname, nil
}

func (i *DNS01Issuer) IssueDNS01(ctx context.Context, hostname string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	hostname, err := normalizeHostname(hostname)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	client, err := i.acmeClient()
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if _, err := client.Register(ctx, &acme.Account{Contact: []string{"mailto:" + i.config.Email}}, acme.AcceptTOS); err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("register acme account: %w", err)
	}
	order, err := solveOrder(ctx, client, []string{hostname}, defaultChallengeRecord, solver)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return i.finalizeOrder(ctx, client, order, []string{hostname})
}

func (i *DNS01Issuer) IssueBundleDNS01(ctx context.Context, hostnames []string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	normalized := make([]string, 0, len(hostnames))
	for _, hostname := range hostnames {
		name, err := normalizeHostname(hostname)
		if err != nil {
			return nil, nil, time.Time{}, err
		}
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return nil, nil, time.Time{}, fmt.Errorf("hostnames are required")
	}
	client, err := i.acmeClient()
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if _, err := client.Register(ctx, &acme.Account{Contact: []string{"mailto:" + i.config.Email}}, acme.AcceptTOS); err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("register acme account: %w", err)
	}
	order, err := solveOrder(ctx, client, normalized, defaultChallengeRecord, solver)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return i.finalizeOrder(ctx, client, order, normalized)
}

func (i *DNS01Issuer) IssueDelegatedDNS01(ctx context.Context, hostname, target string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	hostname, err := normalizeHostname(hostname)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	target = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(target, ".")))
	if target == "" {
		return nil, nil, time.Time{}, fmt.Errorf("delegation target is required")
	}
	client, err := i.acmeClient()
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if _, err := client.Register(ctx, &acme.Account{Contact: []string{"mailto:" + i.config.Email}}, acme.AcceptTOS); err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("register acme account: %w", err)
	}
	order, err := solveOrder(ctx, client, []string{hostname}, func(string) string { return target }, solver)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return i.finalizeOrder(ctx, client, order, []string{hostname})
}

func (i *DNS01Issuer) IssueBundleDelegatedDNS01(ctx context.Context, hostnames []string, target string, solver dns.ChallengeSolver) ([]byte, []byte, time.Time, error) {
	normalized := make([]string, 0, len(hostnames))
	for _, hostname := range hostnames {
		name, err := normalizeHostname(hostname)
		if err != nil {
			return nil, nil, time.Time{}, err
		}
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return nil, nil, time.Time{}, fmt.Errorf("hostnames are required")
	}
	target = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(target, ".")))
	if target == "" {
		return nil, nil, time.Time{}, fmt.Errorf("delegation target is required")
	}
	client, err := i.acmeClient()
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if _, err := client.Register(ctx, &acme.Account{Contact: []string{"mailto:" + i.config.Email}}, acme.AcceptTOS); err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("register acme account: %w", err)
	}
	order, err := solveOrder(ctx, client, normalized, func(string) string { return target }, solver)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return i.finalizeOrder(ctx, client, order, normalized)
}

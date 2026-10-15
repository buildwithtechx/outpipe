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
	"net"
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

func waitForDNSRecord(ctx context.Context, fqdn, value string) error {
	deadline := time.Now().Add(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		if txts, err := net.DefaultResolver.LookupTXT(ctx, fqdn); err == nil {
			for _, txt := range txts {
				if txt == value {
					return nil
				}
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for dns challenge record %s", fqdn)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

type pendingChallenge struct {
	record    string
	digest    string
	challenge *acme.Challenge
}

func defaultChallengeRecord(authzDomain string) string {
	return "_acme-challenge." + authzDomain
}

func solveOrder(ctx context.Context, client *acme.Client, hostnames []string, recordFor func(string) string, solver dns.ChallengeSolver) (*acme.Order, error) {
	if len(hostnames) == 0 {
		return nil, fmt.Errorf("hostnames are required")
	}
	if solver == nil {
		return nil, fmt.Errorf("dns challenge solver is required")
	}
	if len(hostnames) > 1 {
		if _, ok := solver.(dns.ValueCleanupSolver); !ok {
			return nil, fmt.Errorf("bundle issuance needs value-targeted dns cleanup")
		}
	}
	ids := make([]acme.AuthzID, 0, len(hostnames))
	for _, hostname := range hostnames {
		ids = append(ids, acme.AuthzID{Type: "dns", Value: hostname})
	}
	order, err := client.AuthorizeOrder(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("authorize certificate order: %w", err)
	}
	presented := []pendingChallenge{}
	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, authzURL)
		if err != nil {
			cleanupChallenges(ctx, solver, presented)
			return nil, fmt.Errorf("fetch authorization: %w", err)
		}
		for _, challenge := range authz.Challenges {
			if challenge.Type != "dns-01" || challenge.Status != acme.StatusPending {
				continue
			}
			digest, err := client.DNS01ChallengeRecord(challenge.Token)
			if err != nil {
				cleanupChallenges(ctx, solver, presented)
				return nil, fmt.Errorf("compute dns challenge: %w", err)
			}
			record := recordFor(authz.Identifier.Value)
			if err := solver.Present(ctx, record, digest); err != nil {
				cleanupChallenges(ctx, solver, presented)
				return nil, fmt.Errorf("publish dns challenge: %w", err)
			}
			presented = append(presented, pendingChallenge{record: record, digest: digest, challenge: challenge})
		}
	}
	for _, pending := range presented {
		if err := waitForDNSRecord(ctx, pending.record, pending.digest); err != nil {
			cleanupChallenges(ctx, solver, presented)
			return nil, err
		}
	}
	for _, pending := range presented {
		accepted, err := client.Accept(ctx, pending.challenge)
		if err != nil {
			cleanupChallenges(ctx, solver, presented)
			return nil, fmt.Errorf("accept dns challenge: %w", err)
		}
		if _, err := client.WaitAuthorization(ctx, accepted.URI); err != nil {
			cleanupChallenges(ctx, solver, presented)
			return nil, fmt.Errorf("validate dns challenge: %w", err)
		}
	}
	cleanupChallenges(ctx, solver, presented)
	return order, nil
}

func cleanupChallenges(ctx context.Context, solver dns.ChallengeSolver, presented []pendingChallenge) {
	for _, pending := range presented {
		if cleaner, ok := solver.(dns.ValueCleanupSolver); ok {
			_ = cleaner.CleanupValue(ctx, pending.record, pending.digest)
			continue
		}
		_ = solver.Cleanup(ctx, pending.record)
	}
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

func (i *DNS01Issuer) finalizeOrder(ctx context.Context, client *acme.Client, order *acme.Order, hostnames []string) ([]byte, []byte, time.Time, error) {
	ready, err := client.WaitOrder(ctx, order.URI)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("wait for certificate order: %w", err)
	}
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("generate certificate key: %w", err)
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{DNSNames: hostnames}, certKey)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("create certificate request: %w", err)
	}
	chain, _, err := client.CreateOrderCert(ctx, ready.FinalizeURL, csr, true)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("finalize certificate order: %w", err)
	}
	if len(chain) == 0 {
		return nil, nil, time.Time{}, fmt.Errorf("acme returned an empty certificate chain")
	}
	leaf, err := x509.ParseCertificate(chain[0])
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("parse issued certificate: %w", err)
	}
	certPEM := make([]byte, 0, len(chain)*2048)
	for _, der := range chain {
		certPEM = append(certPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})...)
	}
	keyDER, err := x509.MarshalECPrivateKey(certKey)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("encode certificate key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	for _, hostname := range hostnames {
		certFile, keyFile := HostCertificatePaths(i.config.CacheDir, hostname)
		if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("store issued certificate: %w", err)
		}
		if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("store issued certificate key: %w", err)
		}
	}
	return certPEM, keyPEM, leaf.NotAfter, nil
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

package certificates

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/acme"
	"outpipe.dev/outpipe/internal/infra/dns"
)

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

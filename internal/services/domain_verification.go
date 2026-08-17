package services

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/validation"
)

type DomainVerificationService struct {
	domains    *DomainService
	httpClient *http.Client
}

func NewDomainVerificationService(domains *DomainService) (*DomainVerificationService, error) {

	if domains == nil {
		return nil, fmt.Errorf("domain service is required")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {

			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}

			return validation.ValidateSafeTarget(req.URL.Host)
		},
	}
	return &DomainVerificationService{domains: domains, httpClient: client}, nil
}

func (s *DomainVerificationService) VerifyOwnership(ctx context.Context, domain *models.Domain, rawToken string) (bool, error) {

	if domain == nil {
		return false, fmt.Errorf("domain is required")
	}

	if err := validation.ValidateSafeTarget(domain.Hostname); err != nil {
		return false, fmt.Errorf("domain verification target is invalid: %w", err)
	}

	verified := false

	switch strings.ToLower(domain.VerificationMethod) {
	case "dns":
		txts, err := net.LookupTXT("_outpipe-challenge." + domain.Hostname)

		if err == nil {

			for _, txt := range txts {

				if auth.EqualHash(domain.VerificationToken, strings.TrimSpace(txt)) || txt == rawToken {
					verified = true
					break
				}
			}
		}
	case "http":
		reqURL := "http://" + domain.Hostname + "/.well-known/outpipe-challenge"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)

		if err == nil {
			resp, err := s.httpClient.Do(req)

			if err == nil {
				defer resp.Body.Close()
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
				content := strings.TrimSpace(string(body))

				if auth.EqualHash(domain.VerificationToken, content) || content == rawToken {
					verified = true
				}
			}
		}
	default:
		return false, fmt.Errorf("unsupported verification method %q", domain.VerificationMethod)
	}

	if !verified {
		return false, fmt.Errorf("custom domain verification failed for %s", domain.Hostname)
	}

	domain.Status = models.DomainStatusVerified

	if s.domains.issuer != nil {
		expires, err := s.domains.issuer.Issue(ctx, domain.Hostname)

		if err != nil {
			return false, fmt.Errorf("issue domain certificate: %w", err)
		}

		domain.Status = models.DomainStatusActive
		domain.CertificateStatus = "ready"
		domain.CertificateExpiresAt = &expires
	}

	now := time.Now()
	domain.VerifiedAt = &now

	if err := s.domains.domains.Update(ctx, domain); err != nil {
		return false, fmt.Errorf("update domain status: %w", err)
	}

	return true, nil
}

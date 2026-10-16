package services

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/validation"
)

func (s *DomainService) Verify(ctx context.Context, id, token string) error {
	domain, err := s.domains.FindByID(ctx, id)

	if err != nil {
		return fmt.Errorf("find domain: %w", err)
	}

	if strings.HasPrefix(domain.Hostname, "*.") && domain.VerificationMethod != DomainMethodDNSAutomation && domain.VerificationMethod != DomainMethodDNSDelegation {
		return fmt.Errorf("wildcard domains require dns-automation verification")
	}

	switch domain.VerificationMethod {
	case DomainMethodDNSAutomation, DomainMethodDNSDelegation:
		return s.verifyAutomated(ctx, &domain)
	case DomainMethodDNS, DomainMethodHTTP:
		return s.verifyManual(ctx, &domain, token)
	default:
		return fmt.Errorf("unsupported verification method %q", domain.VerificationMethod)
	}
}

func (s *DomainService) verifyManual(ctx context.Context, domain *models.Domain, token string) error {
	if err := validation.ValidateSafeTarget(strings.TrimPrefix(domain.Hostname, "*.")); err != nil {
		return fmt.Errorf("domain verification target is invalid: %w", err)
	}

	verified := false

	switch domain.VerificationMethod {
	case DomainMethodDNS:
		txts, err := net.DefaultResolver.LookupTXT(ctx, "_outpipe-challenge."+domain.Hostname)

		if err == nil {
			for _, txt := range txts {
				if auth.EqualHash(domain.VerificationToken, strings.TrimSpace(txt)) {
					verified = true
					break
				}
			}
		}
	case DomainMethodHTTP:
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+domain.Hostname+"/.well-known/outpipe-challenge", nil)

		if err == nil {
			response, err := s.httpClient.Do(request)

			if err == nil {
				body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
				_ = response.Body.Close()
				content := strings.TrimSpace(string(body))

				if auth.EqualHash(domain.VerificationToken, content) {
					verified = true
				}
			}
		}
	}

	if !verified {
		return fmt.Errorf("custom domain verification failed for %s", domain.Hostname)
	}

	return s.markVerified(ctx, domain)
}

func (s *DomainService) verifyAutomated(ctx context.Context, domain *models.Domain) error {
	if domain.DelegationTarget != "" {
		if err := s.assertDelegation(ctx, domain); err != nil {
			return err
		}
	}
	expires, err := s.issueAutomation(ctx, domain, []string{domain.Hostname})
	if err != nil {
		return err
	}

	now := s.now()
	domain.Status = models.DomainStatusActive
	domain.VerifiedAt = &now
	domain.CertificateStatus = "ready"
	domain.CertificateExpiresAt = &expires

	if err := s.domains.Update(ctx, domain); err != nil {
		return fmt.Errorf("verify domain: %w", err)
	}

	return nil
}

func (s *DomainService) assertDelegation(ctx context.Context, domain *models.Domain) error {
	canonical, err := net.DefaultResolver.LookupCNAME(ctx, "_acme-challenge."+strings.TrimPrefix(domain.Hostname, "*."))
	if err != nil {
		return fmt.Errorf("delegation record is not published: %w", err)
	}
	expected := strings.TrimSuffix(domain.DelegationTarget, ".") + "."
	if !strings.EqualFold(strings.TrimSpace(canonical), expected) {
		return fmt.Errorf("delegation record does not point at the challenge zone")
	}
	return nil
}

func (s *DomainService) markVerified(ctx context.Context, domain *models.Domain) error {
	now := s.now()
	domain.Status = models.DomainStatusVerified
	domain.VerifiedAt = &now

	if s.issuer != nil {
		expires, err := s.issuer.Issue(ctx, domain.Hostname)

		if err != nil {
			return fmt.Errorf("issue domain certificate: %w", err)
		}

		domain.Status = models.DomainStatusActive
		domain.CertificateStatus = "ready"
		domain.CertificateExpiresAt = &expires
	}

	if err := s.domains.Update(ctx, domain); err != nil {
		return fmt.Errorf("verify domain: %w", err)
	}

	return nil
}

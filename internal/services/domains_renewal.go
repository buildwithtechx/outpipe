package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/infra/dns"
	"outpipe.dev/outpipe/internal/models"
)

func (s *DomainService) RenewDueCertificates(ctx context.Context, within time.Duration) (int, error) {
	before := s.now().Add(within)
	domains, err := s.domains.FindExpiringCertificates(ctx, before)

	if err != nil {
		return 0, err
	}

	renewed := 0
	var failures error
	renewedGroups := map[string]bool{}

	for i := range domains {
		domain := domains[i]

		if domain.BundleGroup != "" {
			if renewedGroups[domain.BundleGroup] {
				continue
			}
			renewedGroups[domain.BundleGroup] = true
			if count, err := s.renewBundle(ctx, domains, domain.BundleGroup); err != nil {
				failures = errors.Join(failures, err)
			} else {
				renewed += count
			}
			continue
		}

		var expires time.Time
		var err error

		if domain.VerificationMethod == DomainMethodDNSAutomation || domain.VerificationMethod == DomainMethodDNSDelegation {
			expires, err = s.renewAutomated(ctx, &domain)
		} else if s.issuer != nil {
			expires, err = s.issuer.Issue(ctx, domain.Hostname)
		} else {
			continue
		}

		if err != nil {
			failures = errors.Join(failures, fmt.Errorf("renew %s: %w", domain.Hostname, err))
			continue
		}

		domain.CertificateExpiresAt = &expires
		domain.CertificateStatus = "ready"

		if err := s.domains.Update(ctx, &domain); err != nil {
			failures = errors.Join(failures, fmt.Errorf("record %s renewal: %w", domain.Hostname, err))
			continue
		}

		renewed++
	}

	return renewed, failures
}

func (s *DomainService) renewBundle(ctx context.Context, domains []models.Domain, group string) (int, error) {
	members := make([]models.Domain, 0, 2)
	for i := range domains {
		if domains[i].BundleGroup == group {
			members = append(members, domains[i])
		}
	}
	if len(members) == 0 {
		return 0, nil
	}
	hostnames := make([]string, 0, len(members))
	for _, member := range members {
		hostnames = append(hostnames, member.Hostname)
	}
	expires, err := s.issueAutomation(ctx, &members[0], hostnames)
	if err != nil {
		return 0, fmt.Errorf("renew bundle %s: %w", group, err)
	}
	for i := range members {
		members[i].CertificateExpiresAt = &expires
		members[i].CertificateStatus = "ready"
		if err := s.domains.Update(ctx, &members[i]); err != nil {
			return 0, fmt.Errorf("record %s renewal: %w", members[i].Hostname, err)
		}
	}
	return len(members), nil
}

func (s *DomainService) issueAutomation(ctx context.Context, domain *models.Domain, hostnames []string) (time.Time, error) {
	if s.dnsIssuer == nil {
		return time.Time{}, fmt.Errorf("automated issuance is not configured")
	}
	if domain.DelegationTarget != "" {
		if s.challenge == nil || s.challengeZone == "" {
			return time.Time{}, fmt.Errorf("delegated issuance is not configured")
		}
		var expires time.Time
		var err error
		if len(hostnames) > 1 {
			_, _, expires, err = s.dnsIssuer.IssueBundleDelegatedDNS01(ctx, hostnames, domain.DelegationTarget, s.challenge)
		} else {
			_, _, expires, err = s.dnsIssuer.IssueDelegatedDNS01(ctx, hostnames[0], domain.DelegationTarget, s.challenge)
		}
		if err != nil {
			return time.Time{}, fmt.Errorf("issue delegated certificate: %w", err)
		}
		return expires, nil
	}
	if s.dnsCredentials == nil {
		return time.Time{}, fmt.Errorf("automated issuance is not configured")
	}
	solver, err := s.dnsSolverForDomain(ctx, domain)
	if err != nil {
		return time.Time{}, err
	}
	var expires time.Time
	if len(hostnames) > 1 {
		_, _, expires, err = s.dnsIssuer.IssueBundleDNS01(ctx, hostnames, solver)
	} else {
		_, _, expires, err = s.dnsIssuer.IssueDNS01(ctx, hostnames[0], solver)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("issue automated certificate: %w", err)
	}
	s.dnsCredentials.TouchUsed(ctx, domain.OrganizationID)
	return expires, nil
}

func (s *DomainService) dnsSolverForDomain(ctx context.Context, domain *models.Domain) (dns.ChallengeSolver, error) {
	if s.dnsCredentials == nil {
		return nil, fmt.Errorf("dns credentials are not configured")
	}

	factory := s.dnsSolver

	if factory == nil {
		factory = SolverFor
	}

	if domain.DNSProvider != "" {
		token, err := s.dnsCredentials.Open(ctx, domain.OrganizationID, domain.DNSProvider)
		if err != nil {
			return nil, err
		}
		return factory(domain.DNSProvider, token)
	}

	credentials, err := s.dnsCredentials.List(ctx, domain.OrganizationID)
	if err != nil {
		return nil, err
	}
	if len(credentials) != 1 {
		return nil, fmt.Errorf("connect exactly one DNS provider to use automated verification")
	}
	token, err := s.dnsCredentials.Open(ctx, domain.OrganizationID, credentials[0].Provider)
	if err != nil {
		return nil, err
	}
	return factory(credentials[0].Provider, token)
}

func (s *DomainService) renewAutomated(ctx context.Context, domain *models.Domain) (time.Time, error) {
	expires, err := s.issueAutomation(ctx, domain, []string{domain.Hostname})
	if err != nil {
		return time.Time{}, fmt.Errorf("issue automated renewal: %w", err)
	}
	return expires, nil
}

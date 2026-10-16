package http

import (
	"context"
	"strings"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/billing"
	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/infra/storage"
	"outpipe.dev/outpipe/internal/infra/telemetry"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func configureBillingService(
	billingRepository repositories.BillingRepository,
	organizations repositories.OrganizationRepository,
	users repositories.UserRepository,
	cfg config.APIConfig,
	metrics *telemetry.MetricsExporter,
) (*services.BillingService, error) {
	billingService, err := services.NewBillingService(billingRepository)
	if err != nil {
		return nil, err
	}
	billingService.SetGracePeriod(cfg.Billing.GracePeriod)
	billingService.SetMetrics(metrics)

	var polarClient *billing.PolarClient
	if cfg.Billing.PolarAccessToken != "" {
		polarClient, err = billing.NewPolar(billing.PolarConfig{
			BaseURL:     cfg.Billing.PolarBaseURL,
			AccessToken: cfg.Billing.PolarAccessToken,
			ProductIDs: map[string]string{
				"link":  cfg.Billing.PolarProductLink,
				"route": cfg.Billing.PolarProductRoute,
				"edge":  cfg.Billing.PolarProductEdge,
			},
			YearlyProductIDs: map[string]string{
				"link":  cfg.Billing.PolarProductLinkYearly,
				"route": cfg.Billing.PolarProductRouteYearly,
				"edge":  cfg.Billing.PolarProductEdgeYearly,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	var paystackClient *billing.PaystackClient
	if cfg.Billing.PaystackSecret != "" {
		paystackClient, err = billing.NewPaystack(billing.PaystackConfig{BaseURL: cfg.Billing.PaystackBaseURL, SecretKey: cfg.Billing.PaystackSecret})
		if err != nil {
			return nil, err
		}
	}

	if polarClient != nil || paystackClient != nil {
		gateway, gatewayErr := billing.NewGateway(billing.GatewayConfig{Polar: polarClient, Paystack: paystackClient, Email: func(ctx context.Context, organizationID string) (string, error) {
			organization, findErr := organizations.FindByID(ctx, organizationID)
			if findErr != nil {
				return "", findErr
			}
			user, findErr := users.FindByID(ctx, organization.OwnerID)
			if findErr != nil {
				return "", findErr
			}
			return user.Email, nil
		}})
		if gatewayErr != nil {
			return nil, gatewayErr
		}
		billingService.SetGateway(gateway)
	}

	billingService.SetNotificationResolver(func(ctx context.Context, organizationID string) (services.BillingNotificationTarget, error) {
		organization, findErr := organizations.FindByID(ctx, organizationID)
		if findErr != nil {
			return services.BillingNotificationTarget{}, findErr
		}
		user, findErr := users.FindByID(ctx, organization.OwnerID)
		if findErr != nil {
			return services.BillingNotificationTarget{}, findErr
		}
		return services.BillingNotificationTarget{Email: user.Email, Name: user.Name, OrganizationName: organization.Name, BillingURL: strings.TrimRight(cfg.App.DashboardURL, "/") + "/organizations/" + organization.Slug + "/billing"}, nil
	}, cfg.App.DashboardURL)

	return billingService, nil
}

func configureDomainService(
	domains repositories.DomainRepository,
	dnsCredentials repositories.DNSProviderRepository,
	cfg config.APIConfig,
) (*services.DomainService, error) {
	domainService, err := services.NewDomainService(domains)
	if err != nil {
		return nil, err
	}

	if cfg.App.ACMEEnabled && cfg.App.ACMEEmail != "" {
		issuer, issuerErr := certificates.NewACMEIssuer(certificates.ACMEConfig{Email: cfg.App.ACMEEmail, Directory: cfg.App.ACMEDirectory, CacheDir: cfg.App.CertificateCache, AllowedHost: func(host string) bool {
			return strings.Contains(host, ".") && !strings.Contains(host, "..")
		}})
		if issuerErr != nil {
			return nil, issuerErr
		}
		domainService.SetCertificateIssuer(issuer)

		dnsIssuer, dnsIssuerErr := certificates.NewDNS01Issuer(certificates.ACMEConfig{Email: cfg.App.ACMEEmail, Directory: cfg.App.ACMEDirectory, CacheDir: cfg.App.CertificateCache})
		if dnsIssuerErr != nil {
			return nil, dnsIssuerErr
		}
		domainService.SetDNS01Issuer(dnsIssuer)
	}

	if strings.TrimSpace(cfg.App.ChallengeZone) != "" && strings.TrimSpace(cfg.App.ChallengeDNSCredential) != "" {
		challengeSolver, challengeErr := services.SolverFor(cfg.App.ChallengeDNSProvider, cfg.App.ChallengeDNSCredential)
		if challengeErr != nil {
			return nil, challengeErr
		}
		domainService.SetChallengeSolver(cfg.App.ChallengeZone, challengeSolver)
	}

	dnsCredentialService, err := services.NewDNSCredentialService(dnsCredentials)
	if err != nil {
		return nil, err
	}
	if cfg.Auth.EncryptionKey != "" {
		secretBox, err := storage.NewSecretBox([]byte(cfg.Auth.EncryptionKey))
		if err != nil {
			return nil, err
		}
		dnsCredentialService.SetProtector(secretBox)
	}
	domainService.SetDNSCredentials(dnsCredentialService)
	domainService.SetCertificateCacheDir(cfg.App.CertificateCache)
	return domainService, nil
}

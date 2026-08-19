package http

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/billing"
	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/infra/mail"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func NewDatabaseDependencies(db *gorm.DB, cfg config.APIConfig) (Dependencies, error) {

	if db == nil {
		return Dependencies{}, fmt.Errorf("database is required")
	}

	users, err := repositories.NewUserRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	adminRepository, err := repositories.NewAdminRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	apiKeyRepository, err := repositories.NewAPIKeyRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	apiKeyService, err := services.NewAPIKeyService(apiKeyRepository)

	if err != nil {
		return Dependencies{}, err
	}

	identities, err := repositories.NewOAuthIdentityRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	sessions, err := repositories.NewSessionRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	deviceLogins, err := repositories.NewDeviceLoginRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	organizations, err := repositories.NewOrganizationRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	invitations, err := repositories.NewOrganizationInvitationRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	tunnels, err := repositories.NewTunnelRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	agents, err := repositories.NewAgentRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	domains, err := repositories.NewDomainRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	billingRepository, err := repositories.NewBillingRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	authService, err := services.NewAuthService(users, identities, sessions, adminRepository, cfg.Auth.SessionTTL)

	if err != nil {
		return Dependencies{}, err
	}

	deviceService, err := services.NewDeviceLoginService(deviceLogins, cfg.Auth.DeviceLoginTTL)

	if err != nil {
		return Dependencies{}, err
	}

	organizationService, err := services.NewOrganizationService(organizations)

	if err != nil {
		return Dependencies{}, err
	}

	accountService, err := services.NewAccountService(users, organizations)

	if err != nil {
		return Dependencies{}, err
	}

	adminService, err := services.NewAdminService(adminRepository)

	if err != nil {
		return Dependencies{}, err
	}

	tunnelService, err := services.NewTunnelService(tunnels)

	if err != nil {
		return Dependencies{}, err
	}

	billingService, err := services.NewBillingService(billingRepository)

	if err != nil {
		return Dependencies{}, err
	}

	billingService.SetGracePeriod(cfg.Billing.GracePeriod)
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
			return Dependencies{}, err
		}
	}

	var paystackClient *billing.PaystackClient

	if cfg.Billing.PaystackSecret != "" {
		paystackClient, err = billing.NewPaystack(billing.PaystackConfig{BaseURL: cfg.Billing.PaystackBaseURL, SecretKey: cfg.Billing.PaystackSecret})

		if err != nil {
			return Dependencies{}, err
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
			return Dependencies{}, gatewayErr
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
	var welcomeMailer services.WelcomeMailer
	var invitationMailer services.OrganizationInvitationMailer

	if cfg.Mail.ZeptoAPIKey != "" {
		zepto, mailErr := mail.NewZeptoClient(mail.Config{URL: cfg.Mail.ZeptoURL, APIKey: cfg.Mail.ZeptoAPIKey, FromAddress: cfg.Mail.FromAddress}, nil)

		if mailErr != nil {
			return Dependencies{}, mailErr
		}

		billingMailer, mailErr := mail.NewBillingMailer(zepto, func(ctx context.Context, organizationID string) (string, error) {
			organization, findErr := organizations.FindByID(ctx, organizationID)

			if findErr != nil {
				return "", findErr
			}

			user, findErr := users.FindByID(ctx, organization.OwnerID)

			if findErr != nil {
				return "", findErr
			}

			return user.Email, nil
		}, cfg.App.DashboardURL)

		if mailErr != nil {
			return Dependencies{}, mailErr
		}

		billingService.SetMailer(billingMailer)
		accountMailer, mailErr := mail.NewAccountMailer(zepto, cfg.App.DashboardURL)

		if mailErr != nil {
			return Dependencies{}, mailErr
		}

		accountService.SetMailer(accountMailer)
		welcomeMailer = accountMailer
		invitationMailer = accountMailer
	}

	invitationService, err := services.NewInvitationService(invitations, organizationService, users, cfg.App.DashboardURL, cfg.Auth.InvitationTTL)

	if err != nil {
		return Dependencies{}, err
	}

	invitationService.SetMailer(invitationMailer)
	domainService, err := services.NewDomainService(domains)

	if err != nil {
		return Dependencies{}, err
	}

	if cfg.App.ACMEEmail != "" {
		issuer, issuerErr := certificates.NewACMEIssuer(certificates.ACMEConfig{Email: cfg.App.ACMEEmail, Directory: cfg.App.ACMEDirectory, CacheDir: cfg.App.CertificateCache, AllowedHost: func(host string) bool {
			return strings.Contains(host, ".") && !strings.Contains(host, "..")
		}})

		if issuerErr != nil {
			return Dependencies{}, issuerErr
		}

		domainService.SetCertificateIssuer(issuer)
	}

	tunnelService.SetBilling(billingService)
	organizationService.SetBilling(billingService)
	domainService.SetBilling(billingService)
	hostnameAllocator, err := services.NewHostnameAllocator(tunnels, cfg.Tunnel.Domain)

	if err != nil {
		return Dependencies{}, err
	}

	tunnelService.SetHostnameAllocator(hostnameAllocator)
	agentService, err := services.NewAgentService(agents)

	if err != nil {
		return Dependencies{}, err
	}

	agentService.SetBilling(billingService)
	usageRepository, err := repositories.NewUsageRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	usageService, err := services.NewUsageService(usageRepository)

	if err != nil {
		return Dependencies{}, err
	}

	auditRepository, err := repositories.NewAuditRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	auditService, err := services.NewAuditService(auditRepository)

	if err != nil {
		return Dependencies{}, err
	}

	webhookRepository, err := repositories.NewWebhookRepository(db)

	if err != nil {
		return Dependencies{}, err
	}

	webhookService, err := services.NewWebhookService(webhookRepository)

	if err != nil {
		return Dependencies{}, err
	}

	tunnelService.SetWebhooks(webhookService)

	return Dependencies{Auth: authService, DeviceLogin: deviceService, Organizations: organizationService, Invitations: invitationService, Tunnels: tunnelService, Agents: agentService, Domains: domainService, Usage: usageService, Billing: billingService, Account: accountService, Admin: adminService, Audit: auditService, APIKeys: apiKeyService, Webhooks: webhookService, WelcomeMailer: welcomeMailer, Ready: func(ctx context.Context) error {
		sqlDB, err := db.DB()

		if err != nil {
			return fmt.Errorf("get database connection: %w", err)
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("ping database: %w", err)
		}

		return nil
	}}, nil
}

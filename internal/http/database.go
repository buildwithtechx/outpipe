package http

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/mail"
	"outpipe.dev/outpipe/internal/infra/telemetry"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func NewDatabaseDependencies(db *gorm.DB, cfg config.APIConfig) (Dependencies, error) {
	if db == nil {
		return Dependencies{}, fmt.Errorf("database is required")
	}
	if strings.EqualFold(cfg.App.Environment, "production") && !cfg.App.ACMEEnabled && (cfg.App.ACMEEmail != "" || cfg.App.ACMEDirectory != "") {
		slog.Warn("ACME issuance is disabled; enable OUTPIPE_ACME_ENABLED to issue and renew managed certificates")
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
	dnsCredentials, err := repositories.NewDNSProviderRepository(db)
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
	emailRepository, err := repositories.NewEmailRepository(db)
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
	capturePolicies, err := repositories.NewCapturePolicyRepository(db)
	if err != nil {
		return Dependencies{}, err
	}
	tunnelService.SetCapturePolicies(capturePolicies)

	metrics := telemetry.NewMetricsExporter()
	billingService, err := configureBillingService(billingRepository, organizations, users, cfg, metrics)
	if err != nil {
		return Dependencies{}, err
	}

	resolveBillingRecipient := func(ctx context.Context, organizationID string) (string, error) {
		organization, findErr := organizations.FindByID(ctx, organizationID)
		if findErr != nil {
			return "", findErr
		}
		user, findErr := users.FindByID(ctx, organization.OwnerID)
		if findErr != nil {
			return "", findErr
		}
		return user.Email, nil
	}
	outboxMailer, err := mail.NewOutboxMailer(emailRepository, resolveBillingRecipient, cfg.App.DashboardURL, cfg.Mail.Support)
	if err != nil {
		return Dependencies{}, err
	}
	billingService.SetMailer(outboxMailer)
	accountService.SetMailer(outboxMailer)
	welcomeMailer := services.WelcomeMailer(outboxMailer)
	invitationMailer := services.OrganizationInvitationMailer(outboxMailer)
	supportService, err := services.NewSupportService(outboxMailer)
	if err != nil {
		return Dependencies{}, err
	}

	invitationService, err := services.NewInvitationService(invitations, organizationService, users, cfg.App.DashboardURL, cfg.Auth.InvitationTTL)
	if err != nil {
		return Dependencies{}, err
	}
	invitationService.SetMailer(invitationMailer)

	domainService, err := configureDomainService(domains, dnsCredentials, cfg)
	if err != nil {
		return Dependencies{}, err
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
	webhookService, err := services.NewWebhookService(webhookRepository, services.WithWebhookMetrics(metrics))
	if err != nil {
		return Dependencies{}, err
	}
	tunnelService.SetWebhooks(webhookService)

	secretService, shareService, uptimeService, obsService, aiService, err := productServices(db, cfg, billingRepository)
	if err != nil {
		return Dependencies{}, err
	}
	fleetRepository, err := repositories.NewMachineFleetRepository(db)
	if err != nil {
		return Dependencies{}, err
	}
	fleetService, err := services.NewMachineFleetService(fleetRepository, secretService, tunnelService)
	if err != nil {
		return Dependencies{}, err
	}
	uptimeRepository, err := repositories.NewUptimeRepository(db)
	if err != nil {
		return Dependencies{}, err
	}
	alertRepository, err := repositories.NewAlertRepository(db)
	if err != nil {
		return Dependencies{}, err
	}
	alertBridge, err := services.NewAlertBridgeService(alertRepository, uptimeRepository, uptimeService)
	if err != nil {
		return Dependencies{}, err
	}
	pipelineHealth, err := services.NewPipelineHealthService(webhookRepository, repositories.NewGormTelemetryExportRepository(db), uptimeRepository)
	if err != nil {
		return Dependencies{}, err
	}
	uptimeService.SetSubscriberMailer(outboxMailer)
	uptimeService.SetSubscriberSigningKey(cfg.Auth.EncryptionKey)
	statusBaseURL := strings.TrimSpace(cfg.App.StatusBaseURL)
	if statusBaseURL == "" {
		statusBaseURL = strings.TrimRight(cfg.App.DashboardURL, "/")
	}
	uptimeService.SetSubscriberBaseURL(statusBaseURL)

	return Dependencies{Auth: authService, DeviceLogin: deviceService, Organizations: organizationService, Invitations: invitationService, Tunnels: tunnelService, Agents: agentService, Domains: domainService, Usage: usageService, Billing: billingService, Account: accountService, Admin: adminService, Audit: auditService, APIKeys: apiKeyService, Webhooks: webhookService, Support: supportService, Secrets: secretService, Shares: shareService, Uptime: uptimeService, Observability: obsService, AIAgent: aiService, MachineFleets: fleetService, Pipelines: pipelineHealth, Alerts: alertBridge, WelcomeMailer: welcomeMailer, Ready: databaseReady(db), Metrics: metrics}, nil
}

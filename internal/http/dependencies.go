package http

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/handlers"
	infraredis "outpipe.dev/outpipe/internal/infra/redis"
	"outpipe.dev/outpipe/internal/services"
)

type Dependencies struct {
	Auth          *services.AuthService
	DeviceLogin   *services.DeviceLoginService
	Organizations *services.OrganizationService
	Invitations   *services.InvitationService
	Tunnels       *services.TunnelService
	Agents        *services.AgentService
	Domains       *services.DomainService
	Usage         *services.UsageService
	Billing       *services.BillingService
	OAuth         *services.OAuthService
	Account       *services.AccountService
	Admin         *services.AdminService
	Audit         *services.AuditService
	APIKeys       *services.APIKeyService
	Webhooks      *services.WebhookService
	Support       *services.SupportService
	WelcomeMailer services.WelcomeMailer
	Ready         func(context.Context) error
	RateLimiter   *infraredis.Client
	PublicAPIURL  string
	DashboardURL  string
}

func (d Dependencies) Validate() error {

	if d.Auth == nil || d.DeviceLogin == nil || d.Organizations == nil || d.Invitations == nil || d.Tunnels == nil || d.Agents == nil || d.Domains == nil || d.Usage == nil || d.Billing == nil || d.Account == nil || d.Admin == nil || d.Audit == nil || d.APIKeys == nil || d.Webhooks == nil || d.Support == nil {
		return fmt.Errorf("http service dependencies are incomplete")
	}

	return nil
}

type Handlers struct {
	Health              *handlers.HealthHandler
	Auth                *handlers.AuthHandler
	Organizations       *handlers.OrganizationHandler
	Invitations         *handlers.InvitationHandler
	Tunnels             *handlers.TunnelHandler
	Agents              *handlers.AgentHandler
	Domains             *handlers.DomainHandler
	Usage               *handlers.UsageHandler
	Billing             *handlers.BillingHandler
	OAuth               *handlers.OAuthHandler
	Account             *handlers.AccountHandler
	Admin               *handlers.AdminHandler
	APIKeys             *handlers.APIKeyHandler
	Webhooks            *handlers.WebhookHandler
	Support             *handlers.SupportHandler
	AuditLogs           *handlers.AuditLogHandler
	auditService        *services.AuditService
	authService         *services.AuthService
	organizationService *services.OrganizationService
	apiKeyService       *services.APIKeyService
	agentService        *services.AgentService
}

func buildHandlers(deps Dependencies, cookie handlers.SessionCookieConfig) (Handlers, error) {

	if err := deps.Validate(); err != nil {
		return Handlers{}, err
	}

	authHandler, err := handlers.NewAuthHandler(deps.Auth, deps.DeviceLogin, cookie.Name, cookie.Secure, cookie.Domain)

	if err != nil {
		return Handlers{}, err
	}

	organizationHandler, err := handlers.NewOrganizationHandler(deps.Organizations)

	if err != nil {
		return Handlers{}, err
	}

	invitationHandler, err := handlers.NewInvitationHandler(deps.Invitations)

	if err != nil {
		return Handlers{}, err
	}

	tunnelHandler, err := handlers.NewTunnelHandler(deps.Tunnels)

	if err != nil {
		return Handlers{}, err
	}

	agentHandler, err := handlers.NewAgentHandler(deps.Agents)

	if err != nil {
		return Handlers{}, err
	}

	domainHandler, err := handlers.NewDomainHandler(deps.Domains)

	if err != nil {
		return Handlers{}, err
	}

	usageHandler, err := handlers.NewUsageHandler(deps.Usage)

	if err != nil {
		return Handlers{}, err
	}

	billingHandler, err := handlers.NewBillingHandler(deps.Billing)

	if err != nil {
		return Handlers{}, err
	}

	var oauthHandler *handlers.OAuthHandler

	if deps.OAuth != nil {
		oauthHandler, err = handlers.NewOAuthHandler(deps.OAuth, deps.PublicAPIURL, deps.DashboardURL, cookie.Name, cookie.Secure, cookie.Domain)

		if err != nil {
			return Handlers{}, err
		}
	}

	accountHandler, err := handlers.NewAccountHandler(deps.Account)

	if err != nil {
		return Handlers{}, err
	}

	adminHandler, err := handlers.NewAdminHandler(deps.Admin)

	if err != nil {
		return Handlers{}, err
	}

	apiKeyHandler, err := handlers.NewAPIKeyHandler(deps.APIKeys)

	if err != nil {
		return Handlers{}, err
	}

	webhookHandler, err := handlers.NewWebhookHandler(deps.Webhooks)

	if err != nil {
		return Handlers{}, err
	}

	auditLogHandler, err := handlers.NewAuditLogHandler(deps.Audit)

	if err != nil {
		return Handlers{}, err
	}

	supportHandler, err := handlers.NewSupportHandler(deps.Support)
	if err != nil {
		return Handlers{}, err
	}

	return Handlers{Health: handlers.NewHealthHandler(deps.Ready), Auth: authHandler, Organizations: organizationHandler, Invitations: invitationHandler, Tunnels: tunnelHandler, Agents: agentHandler, Domains: domainHandler, Usage: usageHandler, Billing: billingHandler, OAuth: oauthHandler, Account: accountHandler, Admin: adminHandler, APIKeys: apiKeyHandler, Webhooks: webhookHandler, Support: supportHandler, AuditLogs: auditLogHandler, authService: deps.Auth, organizationService: deps.Organizations, apiKeyService: deps.APIKeys, auditService: deps.Audit, agentService: deps.Agents}, nil
}

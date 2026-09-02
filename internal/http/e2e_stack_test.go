package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type e2eStack struct {
	app            *fiber.App
	internalApp    *fiber.App
	db             *gorm.DB
	userID         string
	organizationID string
	tunnelID       string
	apiKeys        map[string]string
	internalSecret string
	agents         *services.AgentService
	billing        *services.BillingService
}

func newE2EStack(t *testing.T) *e2eStack {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:e2e-%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.OAuthIdentity{}, &models.Session{}, &models.APIKey{}, &models.DeviceLogin{}, &models.Organization{}, &models.OrganizationMember{}, &models.Tunnel{}, &models.Agent{}, &models.Plan{}, &models.Subscription{}, &models.BillingEvent{}, &models.Invoice{}, &models.WebhookSubscription{}, &models.WebhookDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	users, err := repositories.NewUserRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	identities, err := repositories.NewOAuthIdentityRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	sessions, err := repositories.NewSessionRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	keys, err := repositories.NewAPIKeyRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	deviceLogins, err := repositories.NewDeviceLoginRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	organizations, err := repositories.NewOrganizationRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	tunnels, err := repositories.NewTunnelRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	agents, err := repositories.NewAgentRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	billingRepository, err := repositories.NewBillingRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	authService, err := services.NewAuthService(users, identities, sessions, fakePlatformAdminAuthorizer{}, 720*time.Hour)

	if err != nil {
		t.Fatal(err)
	}

	apiKeyService, err := services.NewAPIKeyService(keys)

	if err != nil {
		t.Fatal(err)
	}

	organizationService, err := services.NewOrganizationService(organizations)

	if err != nil {
		t.Fatal(err)
	}

	tunnelService, err := services.NewTunnelService(tunnels)

	if err != nil {
		t.Fatal(err)
	}

	webhookRepository, err := repositories.NewWebhookRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	webhookService, err := services.NewWebhookService(webhookRepository,
		services.WithWebhookHTTPClient(&nethttp.Client{Timeout: 5 * time.Second}),
		services.WithWebhookURLValidator(validation.ValidateURLSyntax),
		services.WithWebhookSynchronousDelivery(),
	)

	if err != nil {
		t.Fatal(err)
	}

	tunnelService.SetWebhooks(webhookService)

	agentService, err := services.NewAgentService(agents)

	if err != nil {
		t.Fatal(err)
	}

	billingService, err := services.NewBillingService(billingRepository)

	if err != nil {
		t.Fatal(err)
	}

	deviceLoginService, err := services.NewDeviceLoginService(deviceLogins, 10*time.Minute)

	if err != nil {
		t.Fatal(err)
	}

	accountService, err := services.NewAccountService(users, organizations)

	if err != nil {
		t.Fatal(err)
	}

	oauthService, err := services.NewOAuthService(authService, map[string]auth.OAuthProvider{"google": fakeOAuthProvider{}}, newMemoryOAuthStateStore())

	if err != nil {
		t.Fatal(err)
	}

	authHandler, err := handlers.NewAuthHandler(authService, deviceLoginService, "outpipe_session", false, "")

	if err != nil {
		t.Fatal(err)
	}

	organizationHandler, err := handlers.NewOrganizationHandler(organizationService)

	if err != nil {
		t.Fatal(err)
	}

	tunnelHandler, err := handlers.NewTunnelHandler(tunnelService)

	if err != nil {
		t.Fatal(err)
	}

	agentHandler, err := handlers.NewAgentHandler(agentService)

	if err != nil {
		t.Fatal(err)
	}

	billingHandler, err := handlers.NewBillingHandler(billingService)

	if err != nil {
		t.Fatal(err)
	}

	accountHandler, err := handlers.NewAccountHandler(accountService)

	if err != nil {
		t.Fatal(err)
	}

	apiKeyHandler, err := handlers.NewAPIKeyHandler(apiKeyService)

	if err != nil {
		t.Fatal(err)
	}

	webhookHandler, err := handlers.NewWebhookHandler(webhookService)

	if err != nil {
		t.Fatal(err)
	}

	oauthHandler, err := handlers.NewOAuthHandler(oauthService, "https://api.outpipe.test", "https://app.outpipe.test", "outpipe_session", false, "")

	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()

	if err := RegisterRoutes(app, Handlers{
		Health:              handlers.NewHealthHandler(func(context.Context) error { return nil }),
		Auth:                authHandler,
		Organizations:       organizationHandler,
		Tunnels:             tunnelHandler,
		Agents:              agentHandler,
		Billing:             billingHandler,
		OAuth:               oauthHandler,
		Account:             accountHandler,
		APIKeys:             apiKeyHandler,
		Webhooks:            webhookHandler,
		authService:         authService,
		organizationService: organizationService,
		apiKeyService:       apiKeyService,
		agentService:        agentService,
	}, RouterOptions{CookieName: "outpipe_session", BillingWebhookSecret: "webhook-secret-1"}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	internalSecret := "internal-secret-1"
	internalApp := fiber.New()

	if err := RegisterInternalRoutes(internalApp, Handlers{
		Health:       handlers.NewHealthHandler(func(context.Context) error { return nil }),
		Auth:         authHandler,
		Tunnels:      tunnelHandler,
		Agents:       agentHandler,
		authService:  authService,
		agentService: agentService,
	}, RouterOptions{InternalAPISecret: internalSecret}); err != nil {
		t.Fatalf("register internal routes: %v", err)
	}

	user := models.User{Email: "e2e-owner@example.com", Name: "E2E Owner", Status: models.UserStatusActive}

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	organization := models.Organization{Name: "E2E Org", Slug: "e2e-org", OwnerID: user.ID}

	if err := db.Create(&organization).Error; err != nil {
		t.Fatal(err)
	}

	member := models.OrganizationMember{OrganizationID: organization.ID, UserID: user.ID, Role: models.MemberRoleOwner}

	if err := db.Create(&member).Error; err != nil {
		t.Fatal(err)
	}

	tunnel := models.Tunnel{OrganizationID: organization.ID, Name: "e2e-tunnel", Protocol: models.TunnelProtocolHTTP, Status: models.TunnelStatusCreated, TargetHost: "127.0.0.1", TargetPort: 3000, PublicHostname: "e2e-tunnel.outpipe.app"}

	if err := db.Create(&tunnel).Error; err != nil {
		t.Fatal(err)
	}

	raw, _, err := apiKeyService.CreateForOrganization(context.Background(), user.ID, organization.ID, "star", []string{"*"}, nil, "")

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = internalApp.Shutdown()
	})

	return &e2eStack{app: app, internalApp: internalApp, db: db, userID: user.ID, organizationID: organization.ID, tunnelID: tunnel.ID, apiKeys: map[string]string{"star": raw}, internalSecret: internalSecret, agents: agentService, billing: billingService}
}

func seededActivePlan(t *testing.T, stack *e2eStack, key string) models.Plan {
	t.Helper()

	plan := models.Plan{Key: key, Name: key, PriceMinor: 1500, Currency: "USD", BillingInterval: models.BillingIntervalMonth, MaxTunnels: 5, MaxConnections: 100, BandwidthBytes: 100 * 1024 * 1024 * 1024, RetentionDays: 30, Active: true}

	if err := stack.db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}

	return plan
}

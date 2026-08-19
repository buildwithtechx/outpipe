package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

type fakePlatformAdminAuthorizer struct{}

func (fakePlatformAdminAuthorizer) IsPlatformAdmin(context.Context, string) (bool, error) {
	return false, nil
}

type verificationStack struct {
	app            *fiber.App
	organizationID string
	tunnelID       string
	apiKeys        map[string]string
}

func newVerificationStack(t *testing.T) *verificationStack {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:scope-%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.OAuthIdentity{}, &models.Session{}, &models.APIKey{}, &models.DeviceLogin{}, &models.Organization{}, &models.OrganizationMember{}, &models.Tunnel{}); err != nil {
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

	organizations, err := repositories.NewOrganizationRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	tunnels, err := repositories.NewTunnelRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	deviceLogins, err := repositories.NewDeviceLoginRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	auth, err := services.NewAuthService(users, identities, sessions, fakePlatformAdminAuthorizer{}, 720*time.Hour)

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

	deviceLoginService, err := services.NewDeviceLoginService(deviceLogins, 10*time.Minute)

	if err != nil {
		t.Fatal(err)
	}

	authHandler, err := handlers.NewAuthHandler(auth, deviceLoginService, "outpipe_session", false, "")

	if err != nil {
		t.Fatal(err)
	}

	tunnelHandler, err := handlers.NewTunnelHandler(tunnelService)

	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()

	if err := RegisterRoutes(app, Handlers{
		Health:              handlers.NewHealthHandler(func(context.Context) error { return nil }),
		Auth:                authHandler,
		Tunnels:             tunnelHandler,
		authService:         auth,
		apiKeyService:       apiKeyService,
		organizationService: organizationService,
	}, RouterOptions{CookieName: "outpipe_session"}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	user := models.User{Email: "cli-user@example.com", Name: "CLI User", Status: models.UserStatusActive}

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	organization := models.Organization{Name: "CLI Org", Slug: "cli-org", OwnerID: user.ID}

	if err := db.Create(&organization).Error; err != nil {
		t.Fatal(err)
	}

	member := models.OrganizationMember{OrganizationID: organization.ID, UserID: user.ID, Role: models.MemberRoleOwner}

	if err := db.Create(&member).Error; err != nil {
		t.Fatal(err)
	}

	tunnel := models.Tunnel{OrganizationID: organization.ID, Name: "cli-tunnel", Protocol: models.TunnelProtocolHTTP, Status: models.TunnelStatusActive, TargetHost: "127.0.0.1", TargetPort: 3000, PublicHostname: "cli-tunnel.outpipe.app"}

	if err := db.Create(&tunnel).Error; err != nil {
		t.Fatal(err)
	}

	otherOrg := models.Organization{Name: "Other Org", Slug: "other-org", OwnerID: user.ID}

	if err := db.Create(&otherOrg).Error; err != nil {
		t.Fatal(err)
	}

	apiKeys := make(map[string]string)
	scopesByName := map[string][]string{
		"star":              {"*"},
		"org-owner":         {"organization:owner"},
		"org-admin":         {"organization:admin"},
		"org-write":         {"organization:write"},
		"org-read":          {"organization:read"},
		"tunnels-read":      {"tunnels:read"},
		"tunnels-write":     {"tunnels:write"},
		"tunnels-readwrite": {"tunnels:read", "tunnels:write"},
		"domains-only":      {"domains:write"},
		"org-restricted":    {"*"},
		"org-view":          {"organization:view"},
		"orgs-write":        {"organizations:write"},
		"tunnels-owner":     {"tunnels:owner"},
	}

	for name, scopes := range scopesByName {
		organizationID := organization.ID

		if name == "org-restricted" {
			organizationID = otherOrg.ID
		}

		raw, _, err := apiKeyService.CreateForOrganization(context.Background(), user.ID, organizationID, name, scopes, nil)

		if err != nil {
			t.Fatal(err)
		}

		apiKeys[name] = raw
	}

	t.Cleanup(func() { _ = app.Shutdown() })

	return &verificationStack{app: app, organizationID: organization.ID, tunnelID: tunnel.ID, apiKeys: apiKeys}
}

func (s *verificationStack) request(t *testing.T, method, path, withKey string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Accept", "application/json")

	if withKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKeys[withKey])
	}

	response, err := s.app.Test(req, -1)

	if err != nil {
		t.Fatalf("%s %s with key %q: %v", method, path, withKey, err)
	}

	return response
}

func denied(status int) bool {
	return status == fiber.StatusForbidden || status == fiber.StatusUnauthorized
}

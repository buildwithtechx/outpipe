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

	db, err := gorm.Open(sqlite.Open("file:scope-"+fmt.Sprintf("%d", time.Now().UnixNano())+"?mode=memory&cache=shared"), &gorm.Config{})

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

func TestCLIManagementCommandScopeMatrix(t *testing.T) {
	stack := newVerificationStack(t)
	orgPath := "/api/v1/organizations/" + stack.organizationID
	tunnelPath := "/api/v1/tunnels/" + stack.tunnelID

	management := []struct {
		name        string
		method      string
		path        string
		allowScopes []string
		denyScopes  []string
	}{
		{
			name:        "create (outpipe create --organization): organization:write",
			method:      http.MethodPost,
			path:        orgPath + "/tunnels",
			allowScopes: []string{"star", "org-owner", "org-admin", "org-write"},
			denyScopes:  []string{"org-read", "tunnels-read", "tunnels-write", "tunnels-readwrite", "domains-only", "org-restricted"},
		},
		{
			name:        "list (outpipe list --organization): organization:read",
			method:      http.MethodGet,
			path:        orgPath + "/tunnels",
			allowScopes: []string{"star", "org-owner", "org-admin", "org-write", "org-read"},
			denyScopes:  []string{"tunnels-read", "tunnels-write", "tunnels-readwrite", "domains-only", "org-restricted"},
		},
		{
			name:        "inspect (outpipe inspect): tunnels:read, write implies read",
			method:      http.MethodGet,
			path:        tunnelPath,
			allowScopes: []string{"star", "tunnels-read", "tunnels-write", "tunnels-readwrite"},
			denyScopes:  []string{"org-owner", "org-admin", "org-write", "org-read", "domains-only", "org-restricted"},
		},
		{
			name:        "start/stop (outpipe start/stop): tunnels:write",
			method:      http.MethodPatch,
			path:        tunnelPath + "/status",
			allowScopes: []string{"star", "tunnels-write", "tunnels-readwrite"},
			denyScopes:  []string{"org-owner", "org-admin", "org-write", "org-read", "tunnels-read", "domains-only", "org-restricted"},
		},
		{
			name:        "revoke (outpipe revoke): tunnels:write",
			method:      http.MethodDelete,
			path:        tunnelPath,
			allowScopes: []string{"star", "tunnels-write", "tunnels-readwrite"},
			denyScopes:  []string{"org-owner", "org-admin", "org-write", "org-read", "tunnels-read", "domains-only", "org-restricted"},
		},
	}

	for _, route := range management {

		for _, key := range route.allowScopes {
			response := stack.request(t, route.method, route.path, key)
			status := response.StatusCode

			if denied(status) {
				t.Errorf("%s: key %q must be allowed, got %d", route.name, key, status)
			}

			_ = response.Body.Close()
		}

		for _, key := range route.denyScopes {
			response := stack.request(t, route.method, route.path, key)
			status := response.StatusCode

			if !denied(status) {
				t.Errorf("%s: key %q must be denied, got %d", route.name, key, status)
			}

			_ = response.Body.Close()
		}
	}
}

func TestCLIManagedTunnelResolutionScope(t *testing.T) {
	stack := newVerificationStack(t)
	path := "/api/v1/tunnels/" + stack.tunnelID

	for _, key := range []string{"star", "tunnels-read", "tunnels-write", "tunnels-readwrite"} {
		response := stack.request(t, http.MethodGet, path, key)

		if denied(response.StatusCode) {
			t.Errorf("open --tunnel-id with key %q must be allowed, got %d", key, response.StatusCode)
		}

		_ = response.Body.Close()
	}

	for _, key := range []string{"org-write", "org-restricted", "domains-only"} {
		response := stack.request(t, http.MethodGet, path, key)

		if !denied(response.StatusCode) {
			t.Errorf("open --tunnel-id with key %q must be denied, got %d", key, response.StatusCode)
		}

		_ = response.Body.Close()
	}
}

func TestCLIUnauthenticatedAndInvalidKey(t *testing.T) {
	stack := newVerificationStack(t)

	response := stack.request(t, http.MethodGet, "/api/v1/tunnels/"+stack.tunnelID, "")

	if response.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("missing credentials must be unauthorized, got %d", response.StatusCode)
	}

	_ = response.Body.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tunnels/"+stack.tunnelID, nil)
	req.Header.Set("Authorization", "Bearer not-a-real-key")
	response, err := stack.app.Test(req, -1)

	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("invalid api key must be unauthorized, got %d", response.StatusCode)
	}

	_ = response.Body.Close()
}

func TestCLIHealthIsUnauthenticated(t *testing.T) {
	stack := newVerificationStack(t)

	for _, key := range []string{"", "tunnels-read"} {
		response := stack.request(t, http.MethodGet, "/readyz", key)

		if response.StatusCode != http.StatusOK {
			t.Errorf("readyz with key %q must be public, got %d", key, response.StatusCode)
		}

		_ = response.Body.Close()
	}
}

func TestOrganizationProxyScopeRanks(t *testing.T) {
	stack := newVerificationStack(t)
	path := "/api/v1/organizations/" + stack.organizationID + "/tunnels"
	ranked := []struct {
		key   string
		allow bool
	}{
		{"org-read", true},
		{"org-write", true},
		{"org-admin", true},
		{"org-owner", true},
		{"star", true},
		{"org-view", false},
		{"orgs-write", false},
		{"tunnels-owner", false},
		{"tunnels-read", false},
	}

	for _, entry := range ranked {
		response := stack.request(t, http.MethodGet, path, entry.key)
		got := !denied(response.StatusCode)

		if got != entry.allow {
			t.Errorf("key %q via organization proxy: expected allowed=%v, got status %d", entry.key, entry.allow, response.StatusCode)
		}

		_ = response.Body.Close()
	}
}

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	if response.StatusCode != 401 {
		t.Errorf("missing credentials must be unauthorized, got %d", response.StatusCode)
	}

	_ = response.Body.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tunnels/"+stack.tunnelID, nil)
	req.Header.Set("Authorization", "Bearer not-a-real-key")
	response, err := stack.app.Test(req, -1)

	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != 401 {
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

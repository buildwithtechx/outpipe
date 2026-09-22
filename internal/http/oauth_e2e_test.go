package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestOAuthFlowEndToEnd(t *testing.T) {
	stack := newE2EStack(t)

	start := stack.request(t, stack.app, http.MethodGet, "/api/v1/auth/oauth/google?return_to=/tunnels", nil, "")

	if start.StatusCode != http.StatusFound {
		t.Fatalf("oauth start: status %d, want 302", start.StatusCode)
	}

	location := start.Header.Get("Location")
	parsed, err := url.Parse(location)

	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}

	state := parsed.Query().Get("state")

	if state == "" {
		t.Fatal("authorize url missing state")
	}

	callback := stack.request(t, stack.app, http.MethodGet, "/api/v1/auth/oauth/google/callback?state="+url.QueryEscape(state)+"&code=authorization-code", nil, "")

	if callback.StatusCode != http.StatusFound {
		t.Fatalf("oauth callback: status %d, want 302", callback.StatusCode)
	}

	if got := callback.Header.Get("Location"); got != "https://app.outpipe.test/tunnels" {
		t.Errorf("callback redirect: %q, want dashboard return path", got)
	}

	sessionValue := stack.sessionCookie(t, callback)

	if sessionValue == "" {
		t.Fatal("callback did not set the session cookie")
	}

	account := stack.request(t, stack.app, http.MethodGet, "/api/v1/account", map[string]string{"Cookie": "outpipe_session=" + sessionValue}, "")

	if account.StatusCode != http.StatusOK {
		t.Fatalf("account with session: status %d", account.StatusCode)
	}

	var profile struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(account.Body).Decode(&profile); err != nil {
		t.Fatalf("decode account: %v", err)
	}

	if profile.Email != "e2e@example.com" {
		t.Errorf("account email = %q, want e2e@example.com", profile.Email)
	}

	organizations := stack.request(t, stack.app, http.MethodGet, "/api/v1/organizations", map[string]string{"Cookie": "outpipe_session=" + sessionValue}, "")

	if organizations.StatusCode != http.StatusOK {
		t.Fatalf("organizations with session: status %d", organizations.StatusCode)
	}
}

func TestOAuthStartRejectsAbsoluteReturnPath(t *testing.T) {
	stack := newE2EStack(t)

	for _, path := range []string{"https://evil.example.com/", "//evil.example.com", "https://app.outpipe.test/tunnels"} {
		start := stack.request(t, stack.app, http.MethodGet, "/api/v1/auth/oauth/google?return_to="+url.QueryEscape(path), nil, "")

		if start.StatusCode != http.StatusBadRequest {
			t.Errorf("return path %q: status %d, want 400", path, start.StatusCode)
		}
	}
}

func TestOAuthCallbackRejectsUnknownState(t *testing.T) {
	stack := newE2EStack(t)

	callback := stack.request(t, stack.app, http.MethodGet, "/api/v1/auth/oauth/google/callback?state=unknown-state&code=code", nil, "")

	if callback.StatusCode != http.StatusFound || !strings.Contains(callback.Header.Get("Location"), "error=oauth_failed") {
		t.Errorf("callback with unknown state: status %d location %q, want oauth_failed redirect", callback.StatusCode, callback.Header.Get("Location"))
	}
}

func TestOAuthUnavailableProvider(t *testing.T) {
	stack := newE2EStack(t)

	start := stack.request(t, stack.app, http.MethodGet, "/api/v1/auth/oauth/github", nil, "")

	if start.StatusCode != http.StatusFound || !strings.Contains(start.Header.Get("Location"), "error=oauth_start_failed") {
		t.Errorf("unavailable provider: status %d location %q", start.StatusCode, start.Header.Get("Location"))
	}
}

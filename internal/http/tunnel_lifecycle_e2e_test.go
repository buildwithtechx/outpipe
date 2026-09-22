package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"outpipe.dev/outpipe/internal/models"
)

func TestTunnelLifecycleEndToEnd(t *testing.T) {
	stack := newE2EStack(t)

	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}
	create := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/tunnels", headers, `{"name":"e2e-live","protocol":"http","targetHost":"127.0.0.1","targetPort":8080,"publicHostname":"e2e-live.outpipe.app"}`)

	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create tunnel: status %d", create.StatusCode)
	}

	var created models.Tunnel

	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode created tunnel: %v", err)
	}

	if created.Status != models.TunnelStatusCreated {
		t.Errorf("created status = %q, want created", created.Status)
	}

	inspect := stack.request(t, stack.app, http.MethodGet, "/api/v1/tunnels/"+created.ID, headers, "")

	if inspect.StatusCode != http.StatusOK {
		t.Fatalf("inspect tunnel: status %d", inspect.StatusCode)
	}

	start := stack.request(t, stack.app, http.MethodPatch, "/api/v1/tunnels/"+created.ID+"/status", headers, `{"status":"active"}`)

	if start.StatusCode != http.StatusNoContent {
		t.Fatalf("start tunnel: status %d", start.StatusCode)
	}

	stop := stack.request(t, stack.app, http.MethodPatch, "/api/v1/tunnels/"+created.ID+"/status", headers, `{"status":"disconnected"}`)

	if stop.StatusCode != http.StatusNoContent {
		t.Fatalf("stop tunnel: status %d", stop.StatusCode)
	}

	var stored models.Tunnel

	if err := stack.db.First(&stored, "id = ?", created.ID).Error; err != nil {
		t.Fatal(err)
	}

	if stored.Status != models.TunnelStatusDisconnected {
		t.Errorf("stored status = %q, want disconnected", stored.Status)
	}

	revoke := stack.request(t, stack.app, http.MethodDelete, "/api/v1/tunnels/"+created.ID, headers, "")

	if revoke.StatusCode != http.StatusNoContent {
		t.Fatalf("revoke tunnel: status %d", revoke.StatusCode)
	}

	afterRevoke := stack.request(t, stack.app, http.MethodGet, "/api/v1/tunnels/"+created.ID, headers, "")

	if afterRevoke.StatusCode != http.StatusOK {
		t.Fatalf("inspect revoked tunnel: status %d, want 200", afterRevoke.StatusCode)
	}

	var revoked models.Tunnel

	if err := json.NewDecoder(afterRevoke.Body).Decode(&revoked); err != nil {
		t.Fatalf("decode revoked tunnel: %v", err)
	}

	if revoked.Status != models.TunnelStatusRevoked {
		t.Errorf("revoked status = %q, want revoked", revoked.Status)
	}
}

func TestTunnelLifecycleRequiresAuth(t *testing.T) {
	stack := newE2EStack(t)

	unauthenticated := stack.request(t, stack.app, http.MethodGet, "/api/v1/tunnels/"+stack.tunnelID, nil, "")

	if unauthenticated.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated inspect: status %d, want 401", unauthenticated.StatusCode)
	}

	withKey := stack.request(t, stack.app, http.MethodGet, "/api/v1/tunnels/"+stack.tunnelID, map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}, "")

	if withKey.StatusCode != http.StatusOK {
		t.Fatalf("authenticated inspect: status %d, want 200", withKey.StatusCode)
	}
}

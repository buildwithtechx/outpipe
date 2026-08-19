package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWebhookFailedDeliveryRecorded(t *testing.T) {
	stack := newE2EStack(t)
	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}

	server, received := newWebhookReceiver(t, http.StatusInternalServerError)
	create := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/webhooks", headers, `{"name":"broken","url":"`+server.URL+`"}`)

	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create broken webhook: %d", create.StatusCode)
	}

	var created struct {
		Subscription struct {
			ID string `json:"id"`
		} `json:"subscription"`
	}

	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode broken webhook: %v", err)
	}

	disconnect := stack.request(t, stack.app, http.MethodPatch, "/api/v1/tunnels/"+stack.tunnelID+"/status", headers, `{"status":"disconnected"}`)

	if disconnect.StatusCode != http.StatusNoContent {
		t.Fatalf("set status disconnected: %d", disconnect.StatusCode)
	}

	assertNoWebhook(t, received)

	deliveries := stack.request(t, stack.app, http.MethodGet, "/api/v1/organizations/"+stack.organizationID+"/webhooks/"+created.Subscription.ID+"/deliveries", headers, "")

	if deliveries.StatusCode != http.StatusOK {
		t.Fatalf("list deliveries: %d", deliveries.StatusCode)
	}

	var deliveryRecords []struct {
		Status    string `json:"status"`
		Attempts  int    `json:"attempts"`
		ErrorJson string `json:"error"`
	}

	if err := json.NewDecoder(deliveries.Body).Decode(&deliveryRecords); err != nil {
		t.Fatalf("decode deliveries: %v", err)
	}

	if len(deliveryRecords) != 1 {
		t.Fatalf("deliveries count = %d, want 1", len(deliveryRecords))
	}

	if deliveryRecords[0].Status != "failed" || deliveryRecords[0].Attempts != 1 {
		t.Errorf("delivery status = %q attempts = %d, want failed/1", deliveryRecords[0].Status, deliveryRecords[0].Attempts)
	}
}

func TestWebhookValidationAndDeletion(t *testing.T) {
	stack := newE2EStack(t)
	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}

	invalid := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/webhooks", headers, `{"name":"bad","url":"not-a-url"}`)

	if invalid.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid url: status %d, want 400", invalid.StatusCode)
	}

	invalidEvent := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/webhooks", headers, `{"name":"bad","url":"https://example.com","events":["tunnel.chaos"]}`)

	if invalidEvent.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid event: status %d, want 400", invalidEvent.StatusCode)
	}

	server, received := newWebhookReceiver(t, http.StatusOK)
	create := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/webhooks", headers, `{"name":"temp","url":"`+server.URL+`"}`)

	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create webhook: %d", create.StatusCode)
	}

	var created struct {
		Subscription struct {
			ID string `json:"id"`
		} `json:"subscription"`
	}

	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode created webhook: %v", err)
	}

	deleted := stack.request(t, stack.app, http.MethodDelete, "/api/v1/organizations/"+stack.organizationID+"/webhooks/"+created.Subscription.ID, headers, "")

	if deleted.StatusCode != http.StatusNoContent {
		t.Fatalf("delete webhook: %d", deleted.StatusCode)
	}

	revoke := stack.request(t, stack.app, http.MethodDelete, "/api/v1/tunnels/"+stack.tunnelID, headers, "")

	if revoke.StatusCode != http.StatusNoContent {
		t.Fatalf("revoke tunnel: %d", revoke.StatusCode)
	}

	assertNoWebhook(t, received)
}

func TestTunnelMetadataAndAPIKeySource(t *testing.T) {
	stack := newE2EStack(t)
	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}

	create := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/tunnels", headers, `{"name":"codedock-app","protocol":"http","targetHost":"127.0.0.1","targetPort":8080,"publicHostname":"codedock-app.outpipe.app","metadata":{"codedock":{"project":"p1","service":"s1"}}}`)

	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create tunnel: %d", create.StatusCode)
	}

	var created struct {
		ID       string `json:"id"`
		Metadata string `json:"metadata"`
	}

	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode created tunnel: %v", err)
	}

	if created.Metadata == "" || !strings.Contains(created.Metadata, "p1") {
		t.Errorf("created tunnel metadata = %q, want codedock project link", created.Metadata)
	}

	inspect := stack.request(t, stack.app, http.MethodGet, "/api/v1/tunnels/"+created.ID, headers, "")

	if inspect.StatusCode != http.StatusOK {
		t.Fatalf("inspect tunnel: %d", inspect.StatusCode)
	}

	var inspected struct {
		Metadata string `json:"metadata"`
	}

	if err := json.NewDecoder(inspect.Body).Decode(&inspected); err != nil {
		t.Fatalf("decode inspected tunnel: %v", err)
	}

	if inspected.Metadata != created.Metadata {
		t.Errorf("inspect metadata = %q, want %q", inspected.Metadata, created.Metadata)
	}

	key := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/api-keys", headers, `{"name":"codedock-deploy","scopes":["tunnels:read","tunnels:write"],"source":"codedock"}`)

	if key.StatusCode != http.StatusCreated {
		t.Fatalf("create api key: %d", key.StatusCode)
	}

	list := stack.request(t, stack.app, http.MethodGet, "/api/v1/organizations/"+stack.organizationID+"/api-keys", headers, "")

	if list.StatusCode != http.StatusOK {
		t.Fatalf("list api keys: %d", list.StatusCode)
	}

	var keys []struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}

	if err := json.NewDecoder(list.Body).Decode(&keys); err != nil {
		t.Fatalf("decode api keys: %v", err)
	}

	found := false

	for _, item := range keys {

		if item.Name == "codedock-deploy" && item.Source == "codedock" {
			found = true
		}
	}

	if !found {
		t.Errorf("api key source not persisted, keys = %+v", keys)
	}
}

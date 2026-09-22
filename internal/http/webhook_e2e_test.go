package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type webhookEnvelope struct {
	ID   string         `json:"id"`
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type capturedWebhook struct {
	request *http.Request
	body    []byte
}

func newWebhookReceiver(t *testing.T, status int) (*httptest.Server, chan capturedWebhook) {
	t.Helper()

	received := make(chan capturedWebhook, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if status >= 200 && status < 300 {
			body, err := io.ReadAll(r.Body)

			if err != nil {
				t.Errorf("read webhook body in receiver: %v", err)
			}

			received <- capturedWebhook{request: r.Clone(r.Context()), body: body}
		}

		w.WriteHeader(status)

		if status >= 200 && status < 300 {
			_, _ = io.WriteString(w, `{"ok":true}`)
		}
	}))

	t.Cleanup(server.Close)
	return server, received
}

func waitForWebhook(t *testing.T, received <-chan capturedWebhook, secret string) webhookEnvelope {
	t.Helper()

	select {
	case captured := <-received:
		var envelope webhookEnvelope

		if err := json.Unmarshal(captured.body, &envelope); err != nil {
			t.Fatalf("decode webhook body: %v", err)
		}

		if envelope.ID == "" || envelope.Type == "" {
			t.Fatalf("webhook envelope missing id or type: %s", captured.body)
		}

		request := captured.request

		if got := request.Header.Get("X-Outpipe-Event-ID"); got != envelope.ID {
			t.Errorf("X-Outpipe-Event-ID = %q, want %q", got, envelope.ID)
		}

		if got := request.Header.Get("X-Outpipe-Event-Type"); got != envelope.Type {
			t.Errorf("X-Outpipe-Event-Type = %q, want %q", got, envelope.Type)
		}

		digest := hmac.New(sha256.New, []byte(secret))
		_, _ = digest.Write(captured.body)
		signature, err := hex.DecodeString(request.Header.Get("X-Outpipe-Signature"))

		if err != nil || !hmac.Equal(digest.Sum(nil), signature) {
			t.Fatalf("invalid webhook signature header %q", request.Header.Get("X-Outpipe-Signature"))
		}

		return envelope

	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for webhook delivery")
		return webhookEnvelope{}
	}
}

func assertNoWebhook(t *testing.T, received <-chan capturedWebhook) {
	t.Helper()

	select {
	case captured := <-received:
		t.Fatalf("unexpected webhook delivery to %s", captured.request.URL)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWebhookSignedDelivery(t *testing.T) {
	stack := newE2EStack(t)
	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}

	server, received := newWebhookReceiver(t, http.StatusOK)
	create := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/webhooks", headers, `{"name":"codedock","url":"`+server.URL+`","events":["tunnel.connected","tunnel.disconnected"]}`)

	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create webhook: status %d", create.StatusCode)
	}

	var created struct {
		Subscription struct {
			ID string `json:"id"`
		} `json:"subscription"`
		Secret string `json:"secret"`
	}

	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create webhook: %v", err)
	}

	if created.Secret == "" || created.Subscription.ID == "" {
		t.Fatal("create webhook response missing subscription id or secret")
	}

	active := stack.request(t, stack.app, http.MethodPatch, "/api/v1/tunnels/"+stack.tunnelID+"/status", headers, `{"status":"active"}`)

	if active.StatusCode != http.StatusNoContent {
		t.Fatalf("set status active: %d", active.StatusCode)
	}

	connected := waitForWebhook(t, received, created.Secret)

	if connected.Type != "tunnel.connected" {
		t.Errorf("event type = %q, want tunnel.connected", connected.Type)
	}

	if tunnelID := connected.Data["tunnel"].(map[string]any)["id"]; tunnelID != stack.tunnelID {
		t.Errorf("event tunnel id = %v, want %s", tunnelID, stack.tunnelID)
	}

	disconnect := stack.request(t, stack.app, http.MethodPatch, "/api/v1/tunnels/"+stack.tunnelID+"/status", headers, `{"status":"disconnected"}`)

	if disconnect.StatusCode != http.StatusNoContent {
		t.Fatalf("set status disconnected: %d", disconnect.StatusCode)
	}

	if event := waitForWebhook(t, received, created.Secret); event.Type != "tunnel.disconnected" {
		t.Errorf("event type = %q, want tunnel.disconnected", event.Type)
	}

	deliveries := stack.request(t, stack.app, http.MethodGet, "/api/v1/organizations/"+stack.organizationID+"/webhooks/"+created.Subscription.ID+"/deliveries", headers, "")

	if deliveries.StatusCode != http.StatusOK {
		t.Fatalf("list deliveries: %d", deliveries.StatusCode)
	}

	var deliveryRecords []struct {
		Status      string    `json:"status"`
		EventType   string    `json:"eventType"`
		DeliveredAt time.Time `json:"deliveredAt"`
	}

	if err := json.NewDecoder(deliveries.Body).Decode(&deliveryRecords); err != nil {
		t.Fatalf("decode deliveries: %v", err)
	}

	if len(deliveryRecords) != 2 {
		t.Fatalf("deliveries count = %d, want 2", len(deliveryRecords))
	}

	for _, delivery := range deliveryRecords {

		if delivery.Status != "delivered" || delivery.DeliveredAt.IsZero() {
			t.Errorf("delivery status = %q deliveredAt zero = %v", delivery.Status, delivery.DeliveredAt.IsZero())
		}
	}
}

func TestWebhookEventFilteringAndRevoke(t *testing.T) {
	stack := newE2EStack(t)
	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}

	server, received := newWebhookReceiver(t, http.StatusOK)

	createRevoked := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/webhooks", headers, `{"name":"revoke-only","url":"`+server.URL+`","events":["tunnel.revoked"]}`)

	if createRevoked.StatusCode != http.StatusCreated {
		t.Fatalf("create revoke-only webhook: %d", createRevoked.StatusCode)
	}

	var revokedOnly struct {
		Secret string `json:"secret"`
	}

	if err := json.NewDecoder(createRevoked.Body).Decode(&revokedOnly); err != nil {
		t.Fatalf("decode revoke-only webhook: %v", err)
	}

	revoke := stack.request(t, stack.app, http.MethodDelete, "/api/v1/tunnels/"+stack.tunnelID, headers, "")

	if revoke.StatusCode != http.StatusNoContent {
		t.Fatalf("revoke tunnel: %d", revoke.StatusCode)
	}

	if event := waitForWebhook(t, received, revokedOnly.Secret); event.Type != "tunnel.revoked" {
		t.Errorf("event type = %q, want tunnel.revoked", event.Type)
	}
}

package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

const webhookSecret = "webhook-secret-1"

func signedPayload(secret, eventID string, body []byte) map[string]string {
	digest := hmac.New(sha256.New, []byte(secret))
	_, _ = digest.Write(body)
	return map[string]string{
		"X-Signature":  hex.EncodeToString(digest.Sum(nil)),
		"X-Event-ID":   eventID,
		"X-Event-Type": "subscription.activated",
	}
}

func TestBillingWebhookTransitionsSubscriptionEndToEnd(t *testing.T) {
	stack := newE2EStack(t)

	plan := models.Plan{Key: "route", Name: "Route", PriceMinor: 1500, Currency: "USD", BillingInterval: models.BillingIntervalMonth, MaxTunnels: 5, MaxConnections: 100, BandwidthBytes: 100 * 1024 * 1024 * 1024, RetentionDays: 30, Active: true}

	if err := stack.db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}

	subscription := models.Subscription{OrganizationID: stack.organizationID, PlanID: plan.ID, Provider: models.BillingProviderPolar, ProviderSubID: "sub_1", Status: models.SubscriptionStatusCanceled, BillingInterval: models.BillingIntervalMonth}

	if err := stack.db.Create(&subscription).Error; err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"id":"evt-activate-1","type":"subscription.activated","data":{"subscription_id":"sub_1","status":"active","current_period_end":"2027-01-01T00:00:00Z","metadata":{"billing_interval":"year"}}}`)
	response := stack.request(t, stack.app, http.MethodPost, "/api/v1/billing/webhooks/polar", signedPayload(webhookSecret, "evt-activate-1", payload), string(payload))

	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("webhook: status %d, want 202", response.StatusCode)
	}

	var stored models.Subscription

	if err := stack.db.First(&stored, "provider_sub_id = ?", "sub_1").Error; err != nil {
		t.Fatal(err)
	}

	if stored.Status != models.SubscriptionStatusActive {
		t.Errorf("subscription status = %q, want active", stored.Status)
	}

	if stored.BillingInterval != models.BillingIntervalYear {
		t.Errorf("subscription interval = %q, want year", stored.BillingInterval)
	}

	if stored.CurrentPeriodEnd == nil || stored.CurrentPeriodEnd.Format(time.RFC3339) != "2027-01-01T00:00:00Z" {
		t.Errorf("subscription period end = %v, want 2027-01-01T00:00:00Z", stored.CurrentPeriodEnd)
	}
}

func TestBillingWebhookIsIdempotent(t *testing.T) {
	stack := newE2EStack(t)

	subscription := models.Subscription{OrganizationID: stack.organizationID, PlanID: "00000000-0000-0000-0000-000000000001", Provider: models.BillingProviderPolar, ProviderSubID: "sub_2", Status: models.SubscriptionStatusCanceled, BillingInterval: models.BillingIntervalMonth}

	if err := stack.db.Create(&subscription).Error; err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"id":"evt-repeat-1","type":"subscription.activated","data":{"status":"active","metadata":{"billing_interval":"month"}}}`)

	for i := 0; i < 2; i++ {
		response := stack.request(t, stack.app, http.MethodPost, "/api/v1/billing/webhooks/polar", signedPayload(webhookSecret, "evt-repeat-1", payload), string(payload))

		if response.StatusCode != http.StatusAccepted {
			t.Fatalf("webhook attempt %d: status %d, want 202", i+1, response.StatusCode)
		}
	}

	var count int64

	if err := stack.db.Model(&models.BillingEvent{}).Where("provider_event_id = ?", "evt-repeat-1").Count(&count).Error; err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("billing events stored = %d, want 1", count)
	}
}

func TestBillingWebhookRejectsBadSignature(t *testing.T) {
	stack := newE2EStack(t)

	payload := []byte(`{"id":"evt-forged-1","type":"subscription.activated","data":{"status":"active"}}`)
	headers := map[string]string{"X-Signature": "deadbeef", "X-Event-ID": "evt-forged-1"}
	response := stack.request(t, stack.app, http.MethodPost, "/api/v1/billing/webhooks/polar", headers, string(payload))

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("forged signature: status %d, want 401", response.StatusCode)
	}
}

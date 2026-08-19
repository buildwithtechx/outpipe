package http

import (
	"context"
	"net/http"
	"testing"

	"outpipe.dev/outpipe/internal/models"
)

func TestAgentHeartbeatAuthenticatesWithAgentCredential(t *testing.T) {
	stack := newE2EStack(t)

	raw, agent, err := stack.agents.Register(context.Background(), stack.organizationID, "edge-agent")

	if err != nil {
		t.Fatal(err)
	}

	headers := map[string]string{"Authorization": "Bearer " + raw, "Content-Type": "application/json"}
	response := stack.request(t, stack.app, http.MethodPost, "/api/v1/agents/"+agent.ID+"/heartbeat", headers, `{"version":"1.0.0","hostname":"edge-1","platform":"linux"}`)

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("heartbeat: status %d, want 204", response.StatusCode)
	}

	var stored models.Agent

	if err := stack.db.First(&stored, "id = ?", agent.ID).Error; err != nil {
		t.Fatal(err)
	}

	if stored.LastSeenAt == nil {
		t.Error("heartbeat did not record last seen at")
	}
}

func TestAgentHeartbeatRejectsInvalidCredentials(t *testing.T) {
	stack := newE2EStack(t)

	raw, agent, err := stack.agents.Register(context.Background(), stack.organizationID, "edge-agent")

	if err != nil {
		t.Fatal(err)
	}

	_ = raw
	headers := map[string]string{"Authorization": "Bearer invalid-token"}
	response := stack.request(t, stack.app, http.MethodPost, "/api/v1/agents/"+agent.ID+"/heartbeat", headers, `{"version":"1.0.0"}`)

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("invalid token: status %d, want 401", response.StatusCode)
	}

	otherRaw, otherAgent, err := stack.agents.Register(context.Background(), stack.organizationID, "other-edge")

	if err != nil {
		t.Fatal(err)
	}

	mismatch := stack.request(t, stack.app, http.MethodPost, "/api/v1/agents/"+agent.ID+"/heartbeat", map[string]string{"Authorization": "Bearer " + otherRaw}, `{"version":"1.0.0"}`)

	if mismatch.StatusCode != http.StatusForbidden {
		t.Fatalf("token mismatch: status %d, want 403", mismatch.StatusCode)
	}

	_ = otherAgent
}

func TestInternalRelayAuthentication(t *testing.T) {
	stack := newE2EStack(t)

	plan := seededActivePlan(t, stack, "route")

	subscription := models.Subscription{OrganizationID: stack.organizationID, PlanID: plan.Key, Provider: models.BillingProviderPolar, ProviderSubID: "sub_relay", Status: models.SubscriptionStatusActive, BillingInterval: models.BillingIntervalMonth}

	if err := stack.db.Create(&subscription).Error; err != nil {
		t.Fatal(err)
	}

	stack.agents.SetBilling(stack.billing)

	raw, agent, err := stack.agents.Register(context.Background(), stack.organizationID, "relay-agent")

	if err != nil {
		t.Fatal(err)
	}

	authenticated := stack.request(t, stack.internalApp, http.MethodGet, "/internal/agents/authenticate", map[string]string{"X-Internal-Secret": stack.internalSecret, "Authorization": "Bearer " + raw}, "")

	if authenticated.StatusCode != http.StatusOK {
		t.Fatalf("internal authenticate: status %d, want 200", authenticated.StatusCode)
	}

	unauthorized := stack.request(t, stack.internalApp, http.MethodGet, "/internal/agents/authenticate", map[string]string{"Authorization": "Bearer " + raw}, "")

	if unauthorized.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing internal secret: status %d, want 401", unauthorized.StatusCode)
	}

	badToken := stack.request(t, stack.internalApp, http.MethodGet, "/internal/agents/authenticate", map[string]string{"X-Internal-Secret": stack.internalSecret, "Authorization": "Bearer forged"}, "")

	if badToken.StatusCode != http.StatusUnauthorized {
		t.Fatalf("forged agent token: status %d, want 401", badToken.StatusCode)
	}

	_ = agent
}

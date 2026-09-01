package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"outpipe.dev/outpipe/internal/models"
)

func TestAgentListEndToEnd(t *testing.T) {
	stack := newE2EStack(t)
	headers := map[string]string{"Authorization": "Bearer " + stack.apiKeys["star"]}

	created := stack.request(t, stack.app, http.MethodPost, "/api/v1/organizations/"+stack.organizationID+"/agents", headers, `{"name":"ci-runner"}`)

	if created.StatusCode != http.StatusCreated {
		t.Fatalf("register agent: status %d", created.StatusCode)
	}

	listed := stack.request(t, stack.app, http.MethodGet, "/api/v1/organizations/"+stack.organizationID+"/agents", headers, "")

	if listed.StatusCode != http.StatusOK {
		t.Fatalf("list agents: status %d", listed.StatusCode)
	}

	var agents []models.Agent

	if err := json.NewDecoder(listed.Body).Decode(&agents); err != nil {
		t.Fatalf("decode agents: %v", err)
	}

	if len(agents) != 1 || agents[0].Name != "ci-runner" {
		t.Fatalf("agents = %#v, want ci-runner", agents)
	}
}

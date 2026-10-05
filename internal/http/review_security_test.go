package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestSecretReadScopeAndTokenCreationScope(t *testing.T) {
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.SecretProject{}, &models.SecretEnvironment{}, &models.SecretEntry{}, &models.SecretVersion{}, &models.SecretMachineToken{}); err != nil {
		t.Fatal(err)
	}
	repo, err := repositories.NewSecretRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewSecretService(repo, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	handler, err := handlers.NewSecretHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	project, err := service.CreateProject(ctx, stack.organizationID, "backend", "Backend", "", stack.userID)
	if err != nil {
		t.Fatal(err)
	}
	environment, err := service.CreateEnvironment(ctx, stack.organizationID, project.ID, "staging", "Staging", stack.userID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetSecret(ctx, stack.organizationID, project.ID, environment.ID, "TOKEN", "value", "", stack.userID); err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Get("/secrets", handler.CLIGetSecrets)
	protected := app.Group("/organizations", sessionRequired(stack.auth, stack.keys, "session"))
	protected.Post("/:organizationID/tokens", organizationRoleRequired(stack.organizations, models.MemberRoleAdmin), handler.CreateMachineToken)
	request := httptest.NewRequest("POST", "/organizations/"+stack.organizationID+"/tokens", strings.NewReader(`{"name":"scoped","projectId":"backend","environmentId":"staging","scopes":["secrets:read"]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+stack.apiKeys["star"])
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 201 {
		t.Fatalf("scope creation returned %d", response.StatusCode)
	}
	var created services.CreatedMachineTokenDTO
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Token.ProjectID == nil || *created.Token.ProjectID != project.ID || created.Token.EnvironmentID == nil || *created.Token.EnvironmentID != environment.ID {
		t.Fatal("route dropped requested scope")
	}
	for _, scopes := range [][]string{nil, {"tunnels:write"}, {"secrets:read"}} {
		token, err := service.CreateMachineToken(ctx, stack.organizationID, "", "", "CI", scopes, stack.userID)
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest("GET", "/secrets?project=backend&environment=staging", nil)
		request.Header.Set("Authorization", "Bearer "+token.Raw)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		expected := 403
		if len(scopes) > 0 && scopes[0] == "secrets:read" {
			expected = 200
		}
		if response.StatusCode != expected {
			t.Fatalf("scopes %v returned %d", scopes, response.StatusCode)
		}
		if expected == 200 {
			var values map[string]string
			if err := json.NewDecoder(response.Body).Decode(&values); err != nil {
				t.Fatal(err)
			}
			if values["TOKEN"] != "value" {
				t.Fatal("slug request did not return secret")
			}
		}
	}
}

func TestShareCreationRejectsSpoofedTenantAndCreator(t *testing.T) {
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.SecretShareLink{}); err != nil {
		t.Fatal(err)
	}
	repo, err := repositories.NewSecretRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewShareService(repo)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := handlers.NewShareHandler(service)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/shares", handler.Create)
	protected := app.Group("/organizations", sessionRequired(stack.auth, stack.keys, "session"))
	protected.Post("/:organizationID/shares", organizationRoleRequired(stack.organizations, models.MemberRoleMember), handler.CreateOrg)
	cases := []struct {
		path, auth, input string
		status            int
	}{
		{"/shares", "", `{"ciphertext":"encrypted","iv":"iv","keyVerifier":"key","organizationId":"victim","createdById":"victim-user"}`, 403},
		{"/shares", "", `{"ciphertext":"encrypted","iv":"iv","keyVerifier":"key"}`, 201},
		{"/organizations/" + stack.organizationID + "/shares", "", `{"ciphertext":"encrypted","iv":"iv","keyVerifier":"key"}`, 401},
		{"/organizations/" + stack.organizationID + "/shares", stack.apiKeys["org-restricted"], `{"ciphertext":"encrypted","iv":"iv","keyVerifier":"key"}`, 403},
		{"/organizations/" + stack.organizationID + "/shares", stack.apiKeys["star"], `{"ciphertext":"encrypted","iv":"iv","keyVerifier":"key","organizationId":"victim","createdById":"victim-user"}`, 201},
	}
	for _, tc := range cases {
		request := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.input))
		request.Header.Set("Content-Type", "application/json")
		if tc.auth != "" {
			request.Header.Set("Authorization", "Bearer "+tc.auth)
		}
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != tc.status {
			t.Fatalf("%s returned %d, expected %d", tc.path, response.StatusCode, tc.status)
		}
	}
	links, err := service.ListOrgShares(context.Background(), stack.organizationID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].CreatedByID == nil || *links[0].CreatedByID != stack.userID {
		t.Fatal("share creator was not authenticated identity")
	}
	if err := service.RevokeOrgShare(context.Background(), "foreign", links[0].ID); err == nil {
		t.Fatal("foreign organization revoked share")
	}
}

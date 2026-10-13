package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestMachineTunnelDerivesTenantAndScopesRelayCredential(t *testing.T) {
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.SecretMachineToken{}); err != nil {
		t.Fatal(err)
	}
	secretRepo, err := repositories.NewSecretRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := services.NewSecretService(secretRepo, "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	tunnelRepo, err := repositories.NewTunnelRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	tunnels, err := services.NewTunnelService(tunnelRepo)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := handlers.NewTunnelHandler(tunnels)
	if err != nil {
		t.Fatal(err)
	}
	handler.SetMachineService(secrets)
	handler.SetMachineSigningKey("test-signing-key")
	app := fiber.New()
	app.Post("/machine", handler.CreateMachine)
	created, err := secrets.CreateMachineToken(context.Background(), stack.organizationID, "", "", "CI", []string{"tunnels:write"}, stack.userID)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/machine", strings.NewReader(`{"name":"machine","protocol":"http","targetHost":"127.0.0.1","targetPort":3000,"publicHostname":"machine.example.com","organizationId":"foreign-org"}`))
	request.Header.Set("Authorization", "Bearer "+created.Raw)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 201 {
		t.Fatalf("machine creation returned %d", response.StatusCode)
	}
	var result struct {
		Tunnel     models.Tunnel `json:"tunnel"`
		RelayToken string        `json:"relayToken"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Tunnel.OrganizationID != stack.organizationID || result.Tunnel.MachineTokenID == nil || *result.Tunnel.MachineTokenID != created.Token.ID || !result.Tunnel.MachineOwned {
		t.Fatal("machine tunnel owner or tenant was not derived from credential")
	}
	claims, err := auth.VerifyRelayToken(result.RelayToken, "test-signing-key")
	if err != nil {
		t.Fatal(err)
	}
	if claims.TunnelID != result.Tunnel.ID || claims.Org != stack.organizationID || claims.MachineTokenID != created.Token.ID || claims.Exp > created.Token.ExpiresAt.Unix() {
		t.Fatal("relay credential escaped machine scope")
	}
	if err := secrets.RevokeMachineToken(context.Background(), stack.organizationID, created.Token.ID); err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest("POST", "/machine", nil)
	request.Header.Set("Authorization", "Bearer "+created.Raw)
	response, err = app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatalf("revoked machine credential returned %d", response.StatusCode)
	}
}

func TestRefreshMachineRelayTokenRenewsScopedCredential(t *testing.T) {
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.SecretMachineToken{}); err != nil {
		t.Fatal(err)
	}
	secretRepo, err := repositories.NewSecretRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := services.NewSecretService(secretRepo, "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	tunnelRepo, err := repositories.NewTunnelRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	tunnels, err := services.NewTunnelService(tunnelRepo)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := handlers.NewTunnelHandler(tunnels)
	if err != nil {
		t.Fatal(err)
	}
	handler.SetMachineService(secrets)
	handler.SetMachineSigningKey("test-signing-key")
	app := fiber.New()
	app.Post("/machine", handler.CreateMachine)
	app.Post("/machine/:tunnelID/refresh", handler.RefreshMachineRelayToken)
	created, err := secrets.CreateMachineToken(context.Background(), stack.organizationID, "", "", "CI", []string{"tunnels:write"}, stack.userID)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/machine", strings.NewReader(`{"name":"machine","protocol":"http","targetHost":"127.0.0.1","targetPort":3000,"publicHostname":"refresh.example.com"}`))
	request.Header.Set("Authorization", "Bearer "+created.Raw)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 201 {
		t.Fatalf("machine creation returned %d", response.StatusCode)
	}
	var createdResult struct {
		Tunnel     models.Tunnel `json:"tunnel"`
		RelayToken string        `json:"relayToken"`
	}
	if err := json.NewDecoder(response.Body).Decode(&createdResult); err != nil {
		t.Fatal(err)
	}
	refresh := func(token, tunnelID string) (int, string, string) {
		req := httptest.NewRequest("POST", "/machine/"+tunnelID+"/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var result struct {
			RelayToken string    `json:"relayToken"`
			ExpiresAt  time.Time `json:"expiresAt"`
		}
		if resp.StatusCode == 200 {
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
		}
		return resp.StatusCode, result.RelayToken, result.ExpiresAt.Format(time.RFC3339)
	}
	status, relayToken, _ := refresh(created.Raw, createdResult.Tunnel.ID)
	if status != 200 || relayToken == "" {
		t.Fatalf("refresh returned status %d with empty credential", status)
	}
	claims, err := auth.VerifyRelayToken(relayToken, "test-signing-key")
	if err != nil {
		t.Fatal(err)
	}
	if claims.TunnelID != createdResult.Tunnel.ID || claims.MachineTokenID != created.Token.ID || claims.Org != stack.organizationID {
		t.Fatal("refreshed credential escaped machine scope")
	}
	other, err := secrets.CreateMachineToken(context.Background(), stack.organizationID, "", "", "other", []string{"tunnels:write"}, stack.userID)
	if err != nil {
		t.Fatal(err)
	}
	if status, _, _ := refresh(other.Raw, createdResult.Tunnel.ID); status != 404 {
		t.Fatalf("foreign machine credential refresh returned %d", status)
	}
	if status, _, _ := refresh(created.Raw, "00000000-0000-0000-0000-000000000000"); status != 404 {
		t.Fatalf("unknown tunnel refresh returned %d", status)
	}
	if err := tunnels.Revoke(context.Background(), createdResult.Tunnel.ID); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := refresh(created.Raw, createdResult.Tunnel.ID); status != 404 {
		t.Fatalf("revoked tunnel refresh returned %d", status)
	}
	if err := secrets.RevokeMachineToken(context.Background(), stack.organizationID, created.Token.ID); err != nil {
		t.Fatal(err)
	}
	if status, _, _ := refresh(created.Raw, createdResult.Tunnel.ID); status != 401 {
		t.Fatalf("revoked machine credential refresh returned %d", status)
	}
}

package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

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

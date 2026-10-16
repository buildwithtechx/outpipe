package http

import (
	"bytes"
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

func newFleetHTTPApp(t *testing.T) (*fiber.App, *services.MachineFleetService) {
	t.Helper()
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.SecretMachineToken{}, &models.Tunnel{}, &models.MachineFleet{}, &models.MachineFleetToken{}, &models.MachineFleetTunnel{}); err != nil {
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
	fleetRepo, err := repositories.NewMachineFleetRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	fleets, err := services.NewMachineFleetService(fleetRepo, secrets, tunnels)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := handlers.NewMachineFleetHandler(fleets)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	org := app.Group("/orgs/:organizationID/fleets", func(c *fiber.Ctx) error {
		if c.Get("Authorization") != "Bearer org-admin" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.Next()
	})
	org.Post("/", handler.Create)
	org.Get("/", handler.List)
	org.Get("/:fleetID", handler.Status)
	org.Delete("/:fleetID", handler.Delete)
	org.Post("/:fleetID/tokens", handler.AttachToken)
	org.Post("/:fleetID/tunnels", handler.BulkCreateTunnels)
	org.Post("/:fleetID/reconcile", handler.Reconcile)
	return app, fleets
}

func fleetHTTP(t *testing.T, app *fiber.App, method, path, body string) (int, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer org-admin")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var result map[string]any
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		raw := new(bytes.Buffer)
		if _, err := raw.ReadFrom(response.Body); err != nil {
			t.Fatal(err)
		}
		if raw.Len() > 0 {
			if err := json.Unmarshal(raw.Bytes(), &result); err != nil {
				t.Fatalf("decode %s %s: %v", method, path, err)
			}
		}
	}
	return response.StatusCode, result
}

func TestFleetLifecycleOverHTTP(t *testing.T) {
	app, fleets := newFleetHTTPApp(t)
	ctx := context.Background()

	if status, result := fleetHTTP(t, app, "POST", "/orgs/org-1/fleets", `{"name":"edge"}`); status != 201 || result["id"] == nil {
		t.Fatalf("create fleet returned %d: %v", status, result)
	}
	if status, _ := fleetHTTP(t, app, "POST", "/orgs/org-1/fleets", `{"name":""}`); status != 400 {
		t.Fatalf("empty fleet name returned %d", status)
	}
	fleet, err := fleets.CreateFleet(ctx, "org-1", "edge-http", "")
	if err != nil {
		t.Fatal(err)
	}
	if status, _ := fleetHTTP(t, app, "POST", "/orgs/org-1/fleets/"+fleet.ID+"/tokens", `{"machineTokenId":"missing"}`); status != 400 {
		t.Fatalf("foreign token attach returned %d", status)
	}
	if status, _ := fleetHTTP(t, app, "POST", "/orgs/org-1/fleets/"+fleet.ID+"/tunnels", `[]`); status != 400 {
		t.Fatalf("empty bulk create returned %d", status)
	}
	if status, result := fleetHTTP(t, app, "GET", "/orgs/org-1/fleets/"+fleet.ID, ""); status != 200 || result["fleet"] == nil {
		t.Fatalf("fleet status returned %d: %v", status, result)
	}
	if status, _ := fleetHTTP(t, app, "GET", "/orgs/org-1/fleets/missing", ""); status != 404 {
		t.Fatalf("missing fleet status returned %d", status)
	}
	if status, result := fleetHTTP(t, app, "POST", "/orgs/org-1/fleets/"+fleet.ID+"/reconcile", ""); status != 200 {
		t.Fatalf("reconcile returned %d: %v", status, result)
	}
	if status, _ := fleetHTTP(t, app, "DELETE", "/orgs/org-1/fleets/"+fleet.ID, ""); status != 204 {
		t.Fatalf("delete fleet returned %d", status)
	}
}

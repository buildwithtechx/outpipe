package http

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestIngestionRequiresScopedExpiringTenantCredential(t *testing.T) {
	stack := newVerificationStack(t)
	app := fiber.New()
	app.Post("/ingest", ingestionRequired(stack.keys, stack.auth, stack.organizations), func(c *fiber.Ctx) error { return c.SendString(c.Locals("ingestionOrganizationID").(string)) })
	hour := time.Now().Add(time.Hour)
	long := time.Now().Add(48 * time.Hour)
	past := time.Now().Add(-time.Hour)
	cases := []struct {
		name     string
		scopes   []string
		expiry   *time.Time
		conflict string
		expected int
	}{
		{"valid", []string{"telemetry:write"}, &hour, "", 200},
		{"conflicting tenant", []string{"telemetry:write"}, &hour, "another-org", 403},
		{"wrong scope", []string{"tunnels:write"}, &hour, "", 403},
		{"nonexpiring", []string{"telemetry:write"}, nil, "", 403},
		{"long lived", []string{"telemetry:write"}, &long, "", 403},
		{"expired", []string{"telemetry:write"}, &past, "", 401},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, _, err := stack.keys.CreateForOrganization(context.Background(), stack.userID, stack.organizationID, tc.name, tc.scopes, tc.expiry, "")
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest("POST", "/ingest", nil)
			request.Header.Set("Authorization", "Bearer "+raw)
			request.Header.Set("X-Organization-Id", tc.conflict)
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, response.StatusCode)
			}
		})
	}
	response, err := app.Test(httptest.NewRequest("POST", "/ingest", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatalf("unauthenticated ingestion returned %d", response.StatusCode)
	}
}

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

type subscriberHTTPMailer struct {
	confirms  int
	incidents int
}

func (m *subscriberHTTPMailer) SendSubscriberConfirm(_ context.Context, _, _, _ string) error {
	m.confirms++
	return nil
}

func (m *subscriberHTTPMailer) SendSubscriberIncident(_ context.Context, _, _, _, _, _, _ string) error {
	m.incidents++
	return nil
}

func newSubscriberHTTPApp(t *testing.T) (*fiber.App, *services.UptimeService, *subscriberHTTPMailer) {
	t.Helper()
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(
		&models.UptimeMonitor{},
		&models.UptimeCheck{},
		&models.UptimeIncident{},
		&models.UptimeIncidentUpdate{},
		&models.UptimeStatusPage{},
		&models.UptimeSubscriber{},
		&models.UptimeSubscriberComponent{},
		&models.UptimeSubscriberNotification{},
		&models.UptimeSubscriptionAttempt{},
	); err != nil {
		t.Fatal(err)
	}
	uptimeRepo, err := repositories.NewUptimeRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	uptime, err := services.NewUptimeService(uptimeRepo)
	if err != nil {
		t.Fatal(err)
	}
	mailer := &subscriberHTTPMailer{}
	uptime.SetSubscriberMailer(mailer)
	uptime.SetSubscriberSigningKey("http-test-signing-key")
	uptime.SetSubscriberBaseURL("https://status.example.test")
	handler, err := handlers.NewUptimeHandler(uptime)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/status/:slug/subscribe", handler.Subscribe)
	app.Post("/status/:slug/confirm", handler.ConfirmSubscription)
	app.Post("/status/:slug/unsubscribe", handler.UnsubscribeWithToken)
	app.Post("/status-pages/:id/confirm", handler.ConfirmSubscriptionByPage)
	app.Post("/status-pages/:id/unsubscribe", handler.UnsubscribeByPage)
	protected := app.Group("/orgs/:organizationID/status-page",
		func(c *fiber.Ctx) error {
			if c.Get("Authorization") != "Bearer org-member" {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
			return c.Next()
		},
		func(c *fiber.Ctx) error {
			page, err := uptime.GetStatusPageByOrg(c.UserContext(), c.Params("organizationID"))
			if err != nil {
				return c.SendStatus(fiber.StatusNotFound)
			}
			c.Locals("statusPage", *page)
			return c.Next()
		},
	)
	protected.Get("/subscribers", handler.ListSubscribers)
	protected.Patch("/subscriptions", handler.UpdateSubscriberSettings)
	protected.Delete("/subscribers/:subscriberID", handler.RemoveSubscriber)
	protected.Post("/subscribers/:subscriberID/unsubscribe", handler.ManagerUnsubscribe)
	protected.Put("/subscribers/:subscriberID/scope", handler.SetSubscriberScope)
	return app, uptime, mailer
}

func subscriberHTTPPost(t *testing.T, app *fiber.App, method, path, body, auth string) (int, map[string]any) {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if auth != "" {
		request.Header.Set("Authorization", "Bearer "+auth)
	}
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
				t.Fatalf("decode %s %s response: %v", method, path, err)
			}
		}
	}
	return response.StatusCode, result
}

func TestSubscriberPublicLifecycleOverHTTP(t *testing.T) {
	app, uptime, mailer := newSubscriberHTTPApp(t)
	ctx := context.Background()
	page, err := uptime.UpsertStatusPage(ctx, "org-sub-http", "sub-http", "Sub HTTP", "", "", true)
	if err != nil {
		t.Fatal(err)
	}

	if status, _ := subscriberHTTPPost(t, app, "POST", "/status/sub-http/subscribe", `{"email":"bad-email"}`, ""); status != 400 {
		t.Fatalf("invalid email signup returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "POST", "/status/sub-http/subscribe", `{}`, ""); status != 400 {
		t.Fatalf("missing email signup returned %d", status)
	}
	if status, result := subscriberHTTPPost(t, app, "POST", "/status/sub-http/subscribe", `{"email":"Reader@Example.com"}`, ""); status != 201 || result["subscribed"] != true {
		t.Fatalf("signup returned %d: %v", status, result)
	}
	if mailer.confirms != 1 {
		t.Fatalf("confirm emails = %d", mailer.confirms)
	}

	stored, err := uptime.ListManagedSubscribers(ctx, page.ID, "", "pending", 25, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Subscribers) != 1 {
		t.Fatalf("pending subscribers = %+v", stored.Subscribers)
	}
	subscriberID := stored.Subscribers[0].ID
	if stored.Subscribers[0].Email != "reader@example.com" {
		t.Fatalf("pending email leaked or mismatch: %+v", stored.Subscribers[0])
	}
	// The confirm token is hashed at rest and single-use, so exercise confirm
	// over HTTP with a second address (the per-email signup limit would
	// otherwise silently swallow the token).
	if err := uptime.RemoveSubscriber(ctx, page.ID, subscriberID); err != nil {
		t.Fatal(err)
	}
	_, pending, err := uptime.RequestSubscription(ctx, page.Slug, "reader-two@example.com", nil, "192.0.2.21")
	if err != nil {
		t.Fatal(err)
	}

	if status, _ := subscriberHTTPPost(t, app, "POST", "/status/sub-http/confirm", `{"token":"bogus"}`, ""); status != 410 {
		t.Fatalf("bogus confirm returned %d", status)
	}
	if status, result := subscriberHTTPPost(t, app, "POST", "/status/sub-http/confirm", `{"token":"`+pending+`"}`, ""); status != 200 || result["confirmed"] != true {
		t.Fatalf("confirm returned %d: %v", status, result)
	}
	if status, _ := subscriberHTTPPost(t, app, "POST", "/status/sub-http/confirm", `{"token":"`+pending+`"}`, ""); status != 410 {
		t.Fatalf("confirm reuse returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "POST", "/status/sub-http/subscribe", `{"email":"reader@example.com"}`, ""); status != 201 {
		t.Fatalf("confirmed re-signup returned %d", status)
	}
	if mailer.confirms != 2 {
		t.Fatal("confirmed re-signup re-sent a confirm email")
	}

	confirmed, err := uptime.ListManagedSubscribers(ctx, page.ID, "", "confirmed", 25, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(confirmed.Subscribers) != 1 {
		t.Fatalf("confirmed subscribers = %+v", confirmed.Subscribers)
	}

	unsubscribeURL := uptime.UnsubscribeLinkFor(confirmed.Subscribers[0].ID, page.Slug)
	raw := unsubscribeURL[strings.LastIndex(unsubscribeURL, "token=")+len("token="):]
	if status, result := subscriberHTTPPost(t, app, "POST", "/status/sub-http/unsubscribe", `{"token":"`+raw+`"}`, ""); status != 200 || result["unsubscribed"] != true {
		t.Fatalf("unsubscribe returned %d: %v", status, result)
	}
	if status, _ := subscriberHTTPPost(t, app, "POST", "/status/sub-http/unsubscribe", `{"token":"`+raw+`"}`, ""); status != 200 {
		t.Fatalf("unsubscribe idempotency returned %d", status)
	}
}

func TestSubscriberManagementAPIsEnforceOrgScope(t *testing.T) {
	app, uptime, _ := newSubscriberHTTPApp(t)
	ctx := context.Background()
	page, err := uptime.UpsertStatusPage(ctx, "org-sub-http", "sub-http", "Sub HTTP", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = uptime.RequestSubscription(ctx, page.Slug, "managed@example.com", nil, "192.0.2.20")
	if err != nil {
		t.Fatal(err)
	}
	managed, err := uptime.ListManagedSubscribers(ctx, page.ID, "", "pending", 25, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(managed.Subscribers) != 1 {
		t.Fatalf("managed subscribers = %+v", managed.Subscribers)
	}
	subscriberID := managed.Subscribers[0].ID

	if status, _ := subscriberHTTPPost(t, app, "GET", "/orgs/org-sub-http/status-page/subscribers", "", ""); status != 401 {
		t.Fatalf("unauthenticated list returned %d", status)
	}
	if status, result := subscriberHTTPPost(t, app, "GET", "/orgs/org-sub-http/status-page/subscribers?q=managed&status=pending", "", "org-member"); status != 200 || result["subscriptionsEnabled"] != true {
		t.Fatalf("list returned %d: %v", status, result)
	}
	if status, _ := subscriberHTTPPost(t, app, "GET", "/orgs/other-org/status-page/subscribers", "", "org-member"); status != 404 {
		t.Fatalf("foreign org list returned %d", status)
	}

	if status, _ := subscriberHTTPPost(t, app, "PUT", "/orgs/org-sub-http/status-page/subscribers/"+subscriberID+"/scope", `{"componentIds":["no-such-monitor"]}`, "org-member"); status != 400 {
		t.Fatalf("foreign scope returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "PUT", "/orgs/org-sub-http/status-page/subscribers/"+subscriberID+"/scope", `{"componentIds":[]}`, "org-member"); status != 200 {
		t.Fatalf("scope clear returned %d", status)
	}
	if status, result := subscriberHTTPPost(t, app, "POST", "/orgs/org-sub-http/status-page/subscribers/"+subscriberID+"/unsubscribe", "", "org-member"); status != 200 || result["status"] != "unsubscribed" {
		t.Fatalf("manager unsubscribe returned %d: %v", status, result)
	}
	if status, _ := subscriberHTTPPost(t, app, "PATCH", "/orgs/org-sub-http/status-page/subscriptions", `{"subscriptionsEnabled":false}`, "org-member"); status != 200 {
		t.Fatalf("settings toggle returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "PATCH", "/orgs/org-sub-http/status-page/subscriptions", `{}`, "org-member"); status != 400 {
		t.Fatalf("malformed settings body returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "PATCH", "/orgs/org-sub-http/status-page/subscriptions", `{"subscriptionsEnabled":false,"extra":true}`, "org-member"); status != 400 {
		t.Fatalf("extra settings key returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "DELETE", "/orgs/org-sub-http/status-page/subscribers/"+subscriberID, "", "org-member"); status != 204 {
		t.Fatalf("remove returned %d", status)
	}
	if status, _ := subscriberHTTPPost(t, app, "DELETE", "/orgs/org-sub-http/status-page/subscribers/"+subscriberID, "", "org-member"); status != 404 {
		t.Fatalf("remove replay returned %d", status)
	}
}

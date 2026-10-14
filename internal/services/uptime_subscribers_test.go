package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type subscriberTestMailer struct {
	confirms  []string
	incidents int
}

func (m *subscriberTestMailer) SendSubscriberConfirm(_ context.Context, email, _, _ string) error {
	m.confirms = append(m.confirms, email)
	return nil
}

func (m *subscriberTestMailer) SendSubscriberIncident(_ context.Context, _, _, _, _, _, _ string) error {
	m.incidents++
	return nil
}

func subscriberTestService(t *testing.T) (*UptimeService, *subscriberTestMailer) {
	t.Helper()
	db := reviewTestDatabase(t,
		&models.UptimeMonitor{},
		&models.UptimeIncident{},
		&models.UptimeIncidentUpdate{},
		&models.UptimeStatusPage{},
		&models.UptimeSubscriber{},
		&models.UptimeSubscriberComponent{},
		&models.UptimeSubscriberNotification{},
		&models.UptimeSubscriptionAttempt{},
	)
	repo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewUptimeService(repo)
	if err != nil {
		t.Fatal(err)
	}
	mailer := &subscriberTestMailer{}
	service.SetSubscriberMailer(mailer)
	service.SetSubscriberSigningKey("test-subscriber-signing-key")
	service.SetSubscriberBaseURL("https://status.example.test")
	return service, mailer
}

func subscriberTestPage(t *testing.T, service *UptimeService, ctx context.Context, orgID string) models.UptimeStatusPage {
	t.Helper()
	page, err := service.UpsertStatusPage(ctx, orgID, orgID+"-status", "Acme Status", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func TestSubscriberDoubleOptInLifecycle(t *testing.T) {
	service, mailer := subscriberTestService(t)
	ctx := context.Background()
	page := subscriberTestPage(t, service, ctx, "org-sub-1")

	subscriberID, token, err := service.RequestSubscription(ctx, page.Slug, "Ops@Example.com", nil, "192.0.2.10")
	if err != nil {
		t.Fatal(err)
	}
	if subscriberID == "" || token == "" {
		t.Fatal("subscription did not return identity and token")
	}
	if len(mailer.confirms) != 1 || mailer.confirms[0] != "ops@example.com" {
		t.Fatalf("confirm email = %v", mailer.confirms)
	}

	stored, err := service.repo.GetSubscriber(ctx, subscriberID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != models.SubscriberStatusPending || stored.TokenHash == "" {
		t.Fatalf("pending subscriber = %+v", stored)
	}

	if _, err := service.ConfirmSubscription(ctx, page.Slug, token); err != nil {
		t.Fatalf("confirm subscription: %v", err)
	}
	confirmed, err := service.repo.GetSubscriber(ctx, subscriberID)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.Status != models.SubscriberStatusConfirmed || !confirmed.Confirmed || confirmed.TokenHash != "" {
		t.Fatalf("confirmed subscriber = %+v", confirmed)
	}

	if _, err := service.ConfirmSubscription(ctx, page.Slug, token); !errors.Is(err, ErrConfirmTokenInvalid) {
		t.Fatalf("reused confirm token = %v", err)
	}

	unsubscribeToken := service.UnsubscribeLinkFor(subscriberID, page.Slug)
	raw := unsubscribeToken[strings.LastIndex(unsubscribeToken, "token=")+len("token="):]
	if _, err := service.UnsubscribeWithToken(ctx, page.Slug, raw); err != nil {
		t.Fatalf("unsubscribe with token: %v", err)
	}
	unsubscribed, err := service.repo.GetSubscriber(ctx, subscriberID)
	if err != nil {
		t.Fatal(err)
	}
	if unsubscribed.Status != models.SubscriberStatusUnsubscribed || unsubscribed.UnsubscribedAt == nil {
		t.Fatalf("unsubscribed subscriber = %+v", unsubscribed)
	}
}

func TestSubscriberTokensRejectCrossPageAndTampering(t *testing.T) {
	service, _ := subscriberTestService(t)
	ctx := context.Background()
	page := subscriberTestPage(t, service, ctx, "org-sub-2")
	other := subscriberTestPage(t, service, ctx, "org-sub-3")

	_, token, err := service.RequestSubscription(ctx, page.Slug, "user@example.com", nil, "192.0.2.11")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmSubscription(ctx, other.Slug, token); !errors.Is(err, ErrConfirmTokenInvalid) {
		t.Fatalf("cross-page confirm = %v", err)
	}
	if _, err := service.ConfirmSubscription(ctx, page.Slug, token+"tampered"); !errors.Is(err, ErrConfirmTokenInvalid) {
		t.Fatalf("tampered confirm = %v", err)
	}
	if _, err := service.UnsubscribeWithToken(ctx, page.Slug, token); !errors.Is(err, ErrUnsubscribeInvalid) {
		t.Fatalf("confirm token accepted as unsubscribe = %v", err)
	}
}

func TestSubscriberSignupRateLimitsWithoutOracle(t *testing.T) {
	service, mailer := subscriberTestService(t)
	ctx := context.Background()
	page := subscriberTestPage(t, service, ctx, "org-sub-4")

	for i := 0; i < maxSignupsPerIPPerHour; i++ {
		if _, _, err := service.RequestSubscription(ctx, page.Slug, "victim@example.com", nil, "192.0.2.12"); err != nil {
			t.Fatal(err)
		}
	}
	before := len(mailer.confirms)
	if _, _, err := service.RequestSubscription(ctx, page.Slug, "victim@example.com", nil, "192.0.2.12"); err != nil {
		t.Fatalf("rate-limited signup must stay silent: %v", err)
	}
	if len(mailer.confirms) != before {
		t.Fatal("rate-limited signup sent another email")
	}

	if _, _, err := service.RequestSubscription(ctx, page.Slug, "victim@example.com", nil, "192.0.2.99"); err != nil {
		t.Fatalf("email-limited re-signup must stay silent: %v", err)
	}
	if len(mailer.confirms) != before {
		t.Fatal("email-limited signup sent another email")
	}
}

func TestSubscriberManagerScopeAndList(t *testing.T) {
	service, _ := subscriberTestService(t)
	ctx := context.Background()
	page := subscriberTestPage(t, service, ctx, "org-sub-5")
	monitor, err := service.CreateMonitor(ctx, "org-sub-5", CreateMonitorInput{Name: "API", URL: "http://example.com", Protocol: "http", IntervalSeconds: 60, TimeoutSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}

	subscriberID, token, err := service.RequestSubscription(ctx, page.Slug, "scoped@example.com", []string{monitor.ID}, "192.0.2.13")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.RequestSubscription(ctx, page.Slug, "bad@example.com", []string{"foreign-monitor"}, "192.0.2.14"); err == nil {
		t.Fatal("foreign component scope was accepted")
	}
	if _, err := service.ConfirmSubscription(ctx, page.Slug, token); err != nil {
		t.Fatal(err)
	}

	result, err := service.ListManagedSubscribers(ctx, page.ID, "scoped", "confirmed", 25, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Subscribers) != 1 || result.Subscribers[0].ID != subscriberID {
		t.Fatalf("managed list = %+v", result.Subscribers)
	}
	if len(result.Subscribers[0].Components) != 1 || result.Subscribers[0].Components[0] != monitor.ID {
		t.Fatalf("subscriber scope = %v", result.Subscribers[0].Components)
	}
	if result.Counts["confirmed"] != 1 || !result.SubscriptionsEnabled {
		t.Fatalf("managed counts = %+v enabled=%v", result.Counts, result.SubscriptionsEnabled)
	}

	view, err := service.SetSubscriberScope(ctx, page.ID, subscriberID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Components) != 0 {
		t.Fatalf("cleared scope = %v", view.Components)
	}

	unsubView, err := service.ManagerUnsubscribe(ctx, page.ID, subscriberID)
	if err != nil {
		t.Fatal(err)
	}
	if unsubView.Status != models.SubscriberStatusUnsubscribed {
		t.Fatalf("manager unsubscribe = %+v", unsubView)
	}

	if err := service.RemoveSubscriber(ctx, page.ID, subscriberID); err != nil {
		t.Fatal(err)
	}
	after, err := service.ListManagedSubscribers(ctx, page.ID, "", "all", 25, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Subscribers) != 0 {
		t.Fatalf("removed subscriber still listed: %+v", after.Subscribers)
	}
}

func TestSubscriberSettingsToggleBlocksSignup(t *testing.T) {
	service, _ := subscriberTestService(t)
	ctx := context.Background()
	page := subscriberTestPage(t, service, ctx, "org-sub-6")

	if _, err := service.UpdateSubscriberSettings(ctx, page.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.RequestSubscription(ctx, page.Slug, "blocked@example.com", nil, "192.0.2.15"); !errors.Is(err, ErrSubscriptionsDisabled) {
		t.Fatalf("disabled subscriptions signup = %v", err)
	}
}

func TestSubscriberIncidentNotifyEnqueuesAndSends(t *testing.T) {
	service, mailer := subscriberTestService(t)
	ctx := context.Background()
	page := subscriberTestPage(t, service, ctx, "org-sub-7")
	monitor, err := service.CreateMonitor(ctx, "org-sub-7", CreateMonitorInput{Name: "API", URL: "http://example.com", Protocol: "http", IntervalSeconds: 60, TimeoutSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}

	_, tokenAll, err := service.RequestSubscription(ctx, page.Slug, "all@example.com", nil, "192.0.2.16")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmSubscription(ctx, page.Slug, tokenAll); err != nil {
		t.Fatal(err)
	}
	_, tokenScoped, err := service.RequestSubscription(ctx, page.Slug, "scoped@example.com", []string{monitor.ID}, "192.0.2.17")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmSubscription(ctx, page.Slug, tokenScoped); err != nil {
		t.Fatal(err)
	}
	_, tokenWide, err := service.RequestSubscription(ctx, page.Slug, "wide@example.com", nil, "192.0.2.18")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmSubscription(ctx, page.Slug, tokenWide); err != nil {
		t.Fatal(err)
	}

	incident, err := service.CreateIncident(ctx, "org-sub-7", CreateIncidentInput{Title: "Outage", Severity: models.IncidentSeverityMajor, Message: "Investigating"})
	if err != nil {
		t.Fatal(err)
	}
	if len(incident.Updates) != 1 {
		t.Fatalf("incident updates = %+v", incident.Updates)
	}
	if err := service.AddIncidentUpdateScoped(ctx, incident.ID, models.IncidentStatusMonitoring, "Fix applied", []string{monitor.ID}); err != nil {
		t.Fatal(err)
	}

	// The incident's initial update broadcasts to all three confirmed
	// subscribers, and the scoped update reaches them too (unscoped
	// subscribers receive every update): 3 + 3.
	before := mailer.incidents
	if err := service.ProcessPendingSubscriberNotifications(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if mailer.incidents-before != 6 {
		t.Fatalf("scoped incident emails = %d", mailer.incidents-before)
	}

	// Unscoped update reaches everyone.
	if err := service.AddIncidentUpdateScoped(ctx, incident.ID, models.IncidentStatusResolved, "Resolved", nil); err != nil {
		t.Fatal(err)
	}
	before = mailer.incidents
	if err := service.ProcessPendingSubscriberNotifications(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if mailer.incidents-before != 3 {
		t.Fatalf("broadcast incident emails = %d", mailer.incidents-before)
	}

	// Second pass is a no-op (durable outbox drained).
	before = mailer.incidents
	if err := service.ProcessPendingSubscriberNotifications(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if mailer.incidents != before {
		t.Fatal("notification outbox replayed sent rows")
	}
}

package services

import (
	"context"
	"testing"

	"outpipe.dev/outpipe/internal/models"
)

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

	before := mailer.incidents
	if err := service.ProcessPendingSubscriberNotifications(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if mailer.incidents-before != 6 {
		t.Fatalf("scoped incident emails = %d", mailer.incidents-before)
	}

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

	before = mailer.incidents
	if err := service.ProcessPendingSubscriberNotifications(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if mailer.incidents != before {
		t.Fatal("notification outbox replayed sent rows")
	}
}

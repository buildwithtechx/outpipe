package services

import (
	"context"
	"testing"

	"outpipe.dev/outpipe/internal/infra/telemetry"
)

type testReporter struct {
	events []telemetry.Event
}

func (r *testReporter) Report(_ context.Context, event telemetry.Event) error {
	r.events = append(r.events, event)
	return nil
}

func TestAlertService(t *testing.T) {
	reporter := &testReporter{}
	alerts, err := NewAlertService(reporter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := alerts.AlertFailedWebhook(context.Background(), "polar", "evt_1", "bad sig"); err != nil {
		t.Fatalf("unexpected alert error: %v", err)
	}

	if err := alerts.AlertRepeatedPaymentFailure(context.Background(), "org_1", "sub_1", 3); err != nil {
		t.Fatalf("unexpected alert error: %v", err)
	}

	if err := alerts.AlertStalePresenceGrowth(context.Background(), "org_1", 5); err != nil {
		t.Fatalf("unexpected alert error: %v", err)
	}

	if err := alerts.AlertQuotaInconsistency(context.Background(), "org_1", 200, 100); err != nil {
		t.Fatalf("unexpected alert error: %v", err)
	}

	if len(reporter.events) != 4 {
		t.Fatalf("expected 4 reported events, got %d", len(reporter.events))
	}

	if reporter.events[0].Name != string(AlertFailedWebhook) {
		t.Errorf("expected %s, got %s", AlertFailedWebhook, reporter.events[0].Name)
	}
}

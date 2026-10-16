package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func setupAlertBridge(t *testing.T) (*AlertBridgeService, repositories.UptimeRepository) {
	t.Helper()
	db := reviewTestDatabase(t,
		&models.UptimeMonitor{},
		&models.UptimeCheck{},
		&models.UptimeIncident{},
		&models.UptimeIncidentUpdate{},
		&models.UptimeStatusPage{},
		&models.UptimeSubscriber{},
		&models.UptimeSubscriberNotification{},
		&models.ObservabilityAlert{},
		&models.AlertIncident{},
	)
	uptimeRepo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	alertRepo, err := repositories.NewAlertRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	uptimeService, err := NewUptimeService(uptimeRepo)
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := NewAlertBridgeService(alertRepo, uptimeRepo, uptimeService)
	if err != nil {
		t.Fatal(err)
	}
	return bridge, uptimeRepo
}

func recordAlertChecks(t *testing.T, repo repositories.UptimeRepository, monitorID string, successes []bool, base time.Time) {
	t.Helper()
	for i, success := range successes {
		check := &models.UptimeCheck{
			ID:         fmt.Sprintf("%s-check-%d-%d", monitorID, base.UnixNano(), i),
			MonitorID:  monitorID,
			StatusCode: 200,
			Success:    success,
			CheckedAt:  base.Add(time.Duration(i) * time.Second),
		}
		if !success {
			check.StatusCode = 500
		}
		if err := repo.RecordCheck(context.Background(), check); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAlertBridgeFiresAndResolvesIncident(t *testing.T) {
	bridge, uptimeRepo := setupAlertBridge(t)
	ctx := context.Background()
	monitor := &models.UptimeMonitor{ID: "monitor-1", OrganizationID: "org", Name: "api", URL: "https://api.example.com", Protocol: "https", IntervalSeconds: 60}
	if err := uptimeRepo.CreateMonitor(ctx, monitor); err != nil {
		t.Fatal(err)
	}
	alert, err := bridge.Create(ctx, "org", CreateAlertInput{
		Name:                 "api down",
		MonitorIDs:           []string{"monitor-1"},
		ConsecutiveFailures:  2,
		ConsecutiveRecovered: 2,
		EvalIntervalSeconds:  60,
		StatusPageTitle:      "API outage",
		StatusPageImpact:     "outage",
	})
	if err != nil {
		t.Fatal(err)
	}
	recordAlertChecks(t, uptimeRepo, "monitor-1", []bool{false, false}, time.Now().UTC())
	if err := bridge.ProcessDueAlerts(ctx, 100); err != nil {
		t.Fatal(err)
	}
	stored, err := bridge.alerts.FindByID(ctx, alert.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != models.AlertStateFiring {
		t.Fatalf("alert state = %s, want firing", stored.State)
	}
	link, err := bridge.alerts.GetOpenIncident(ctx, alert.ID)
	if err != nil {
		t.Fatal(err)
	}
	incident, err := bridge.incidents.repo.GetIncident(ctx, link.IncidentID)
	if err != nil {
		t.Fatal(err)
	}
	if incident.Status != models.IncidentStatusInvestigating || incident.Severity != models.IncidentSeverityMajor || incident.Title != "API outage" {
		t.Fatalf("firing incident incorrect: %+v", incident)
	}
	recordAlertChecks(t, uptimeRepo, "monitor-1", []bool{true, true}, time.Now().UTC().Add(time.Minute))
	stored.NextEvaluationAt = time.Now().UTC().Add(-time.Minute)
	if err := bridge.alerts.Update(ctx, &stored); err != nil {
		t.Fatal(err)
	}
	if err := bridge.ProcessDueAlerts(ctx, 100); err != nil {
		t.Fatal(err)
	}
	stored, err = bridge.alerts.FindByID(ctx, alert.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != models.AlertStateOK {
		t.Fatalf("alert state = %s, want ok", stored.State)
	}
	resolved, err := bridge.incidents.repo.GetIncident(ctx, link.IncidentID)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Status != models.IncidentStatusResolved || resolved.ResolvedAt == nil {
		t.Fatalf("incident was not resolved: %+v", resolved)
	}
	if _, err := bridge.alerts.GetOpenIncident(ctx, alert.ID); err != repositories.ErrNotFound {
		t.Fatalf("incident link was not closed: %v", err)
	}
}

func TestAlertBridgeRequiresOwnedMonitors(t *testing.T) {
	bridge, _ := setupAlertBridge(t)
	ctx := context.Background()
	if _, err := bridge.Create(ctx, "org", CreateAlertInput{Name: "x", MonitorIDs: []string{"foreign-monitor"}}); err == nil {
		t.Fatal("alert accepted a foreign monitor")
	}
	if _, err := bridge.Create(ctx, "org", CreateAlertInput{Name: "x"}); err == nil {
		t.Fatal("alert accepted zero monitors")
	}
	if _, err := bridge.Create(ctx, "org", CreateAlertInput{Name: "", MonitorIDs: []string{"m"}}); err == nil {
		t.Fatal("alert accepted an empty name")
	}
}

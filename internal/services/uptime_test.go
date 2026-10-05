package services

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestUptimeServiceLifecycle(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	err = db.AutoMigrate(
		&models.UptimeMonitor{},
		&models.UptimeCheck{},
		&models.UptimeIncident{},
		&models.UptimeIncidentUpdate{},
		&models.UptimeStatusPage{},
		&models.UptimeSubscriber{},
	)
	if err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	repo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	service, err := NewUptimeService(repo)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	ctx := context.Background()
	orgID := "org_test_123"

	// 1. Create a monitor targeting a local test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()
	service.httpClient = server.Client()
	service.httpClient.Transport = &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}}

	monitor, err := service.CreateMonitor(ctx, orgID, CreateMonitorInput{
		Name:            "API Gateway",
		URL:             "http://example.com",
		Protocol:        "http",
		Method:          "GET",
		IntervalSeconds: 30,
		TimeoutSeconds:  5,
	})
	if err != nil {
		t.Fatalf("create monitor: %v", err)
	}
	if monitor.ID == "" || monitor.Name != "API Gateway" {
		t.Fatalf("unexpected monitor: %+v", monitor)
	}
	if _, err := service.GetOwnedMonitor(ctx, "other-org", monitor.ID); err == nil {
		t.Fatal("foreign organization accessed monitor")
	}

	// 2. Probe the monitor
	check, err := service.Probe(ctx, &monitor)
	if err != nil {
		t.Fatalf("probe monitor: %v", err)
	}
	if !check.Success || check.StatusCode != 200 {
		t.Fatalf("check failed: %+v", check)
	}
	if monitor.Status != models.MonitorStatusUp || monitor.UptimeRatio != 100.0 {
		t.Fatalf("unexpected monitor state: status=%v ratio=%v", monitor.Status, monitor.UptimeRatio)
	}

	// 3. Create an incident
	incident, err := service.CreateIncident(ctx, orgID, CreateIncidentInput{
		Title:    "Elevated error rates on checkout",
		Severity: models.IncidentSeverityMajor,
		Message:  "Investigating database connection pool saturation",
	})
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}
	if incident.Status != models.IncidentStatusInvestigating || len(incident.Updates) != 1 {
		t.Fatalf("unexpected incident: %+v", incident)
	}

	// 4. Update incident status
	err = service.AddIncidentUpdate(ctx, incident.ID, models.IncidentStatusMonitoring, "Fix applied, monitoring metrics")
	if err != nil {
		t.Fatalf("update incident monitoring: %v", err)
	}

	err = service.AddIncidentUpdate(ctx, incident.ID, models.IncidentStatusResolved, "All systems restored to normal")
	if err != nil {
		t.Fatalf("update incident resolved: %v", err)
	}

	// 5. Upsert status page
	page, err := service.UpsertStatusPage(ctx, orgID, "techx-status", "TechX Status", "Public uptime tracker", "status.techx.com", true)
	if err != nil {
		t.Fatalf("upsert status page: %v", err)
	}
	if page.Slug != "techx-status" {
		t.Fatalf("unexpected slug: %s", page.Slug)
	}

	// 6. Get public status data
	statusData, err := service.GetPublicStatusData(ctx, "techx-status")
	if err != nil {
		t.Fatalf("get public status data: %v", err)
	}
	if statusData.OverallStatus != "operational" {
		t.Fatalf("expected operational, got %s", statusData.OverallStatus)
	}
	if len(statusData.Monitors) != 1 || len(statusData.PastIncidents) != 1 || len(statusData.ActiveIncidents) != 0 {
		t.Fatalf("unexpected counts: monitors=%d past=%d active=%d", len(statusData.Monitors), len(statusData.PastIncidents), len(statusData.ActiveIncidents))
	}

	// 7. Subscribe to status page
	err = service.Subscribe(ctx, "techx-status", "admin@techx.com")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	monitor.MaxLatencyMs = 1
	check, err = service.Probe(ctx, &monitor)
	if err != nil {
		t.Fatal(err)
	}
	if check.Success || check.ErrorMessage != "latency assertion failed" {
		t.Fatal("latency assertion did not fail the check")
	}
	if _, err := service.UpsertStatusPage(ctx, orgID, "techx-status", "TechX Status", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetPublicStatusData(ctx, "techx-status"); err == nil {
		t.Fatal("private status page was public")
	}
	if err := service.Subscribe(ctx, "techx-status", "admin@techx.com"); err == nil {
		t.Fatal("private status page accepted a subscriber")
	}
}

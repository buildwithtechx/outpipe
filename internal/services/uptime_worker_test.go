package services

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestUptimeSchedulerClaimsDueMonitorsOnceAndRecordsOutage(t *testing.T) {
	db := reviewTestDatabase(t, &models.UptimeMonitor{}, &models.UptimeCheck{})
	repo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewUptimeService(repo)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) }))
	defer server.Close()
	svc.httpClient.Transport = &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}}
	ctx := context.Background()
	monitor, err := svc.CreateMonitor(ctx, "org", CreateMonitorInput{Name: "API", URL: "http://example.com", Protocol: "http", IntervalSeconds: 10, TimeoutSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	paused := monitor
	paused.ID, paused.Status = "paused", models.MonitorStatusPaused
	if err := repo.CreateMonitor(ctx, &paused); err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	claimed := make(chan []models.UptimeMonitor, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			items, err := repo.ClaimDueMonitors(ctx, time.Now().UTC(), 16)
			if err != nil {
				t.Error(err)
			}
			claimed <- items
		}()
	}
	wait.Wait()
	close(claimed)
	count := 0
	for items := range claimed {
		for _, item := range items {
			count++
			if item.ID != monitor.ID {
				t.Fatal("paused monitor claimed")
			}
		}
	}
	if count != 1 {
		t.Fatalf("claim count = %d", count)
	}
	if err := db.Model(&models.UptimeMonitor{}).Where("id = ?", monitor.ID).Update("next_probe_at", time.Now().UTC().Add(-time.Second)).Error; err != nil {
		t.Fatal(err)
	}
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- svc.RunScheduler(workerCtx, func(err error) { t.Error(err) }) }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		stored, err := repo.GetMonitor(ctx, monitor.ID)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		if stored.LastCheckAt != nil && stored.Status == models.MonitorStatusDown {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	checks, err := repo.GetRecentChecks(ctx, monitor.ID, 10)
	if err != nil || len(checks) != 1 || checks[0].Success {
		t.Fatalf("scheduled outage not recorded: %v", err)
	}
	if items, err := repo.ClaimDueMonitors(ctx, time.Now().UTC(), 16); err != nil || len(items) != 0 {
		t.Fatalf("monitor ran before interval: %v", err)
	}
	if items, err := repo.ClaimDueMonitors(ctx, time.Now().Add(11*time.Second).UTC(), 16); err != nil || len(items) != 1 {
		t.Fatalf("monitor not due after interval: %v", err)
	}
}

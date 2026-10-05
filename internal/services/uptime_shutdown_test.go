package services

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestUptimeSchedulerFinishesClaimedProbeOnShutdown(t *testing.T) {
	db := reviewTestDatabase(t, &models.UptimeMonitor{}, &models.UptimeCheck{})
	repo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewUptimeService(repo)
	if err != nil {
		t.Fatal(err)
	}
	started, finish := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-finish; w.WriteHeader(200) }))
	defer server.Close()
	svc.httpClient.Transport = &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}}
	ctx := context.Background()
	monitor, err := svc.CreateMonitor(ctx, "org", CreateMonitorInput{Name: "Drain", URL: " http://93.184.216.34 ", Protocol: "http", IntervalSeconds: 10, TimeoutSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- svc.RunScheduler(workerCtx, func(err error) { t.Error(err) }) }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		close(finish)
		t.Fatal("probe did not start")
	}
	cancel()
	close(finish)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("scheduler did not drain")
	}
	checks, err := repo.GetRecentChecks(ctx, monitor.ID, 10)
	if err != nil || len(checks) != 1 || !checks[0].Success {
		t.Fatalf("shutdown lost claimed probe: %v", err)
	}
	stored, err := repo.GetMonitor(ctx, monitor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.NextProbeAt == nil || stored.NextProbeAt.After(time.Now().Add(15*time.Second)) {
		t.Fatal("shutdown retained claim lease")
	}
}

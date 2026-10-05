package services

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"outpipe.dev/outpipe/internal/models"
	"testing"
)

func TestHTTPProbeAssertions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	svc := &UptimeService{httpClient: server.Client()}
	svc.httpClient.Transport = &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}}
	monitor := &models.UptimeMonitor{Protocol: "http", Method: "GET", URL: "http://example.com", ExpectedStatusCode: 201, BodyRegex: `"status":"healthy"`}
	if _, err := svc.probeTarget(context.Background(), monitor); err != nil {
		t.Fatal(err)
	}
	monitor.ExpectedStatusCode = 200
	if _, err := svc.probeTarget(context.Background(), monitor); err == nil {
		t.Fatal("status mismatch accepted")
	}
	monitor.ExpectedStatusCode = 201
	monitor.BodyRegex = "unhealthy"
	if _, err := svc.probeTarget(context.Background(), monitor); err == nil {
		t.Fatal("body mismatch accepted")
	}
	monitor.BodyRegex = "["
	if _, err := svc.probeTarget(context.Background(), monitor); err == nil {
		t.Fatal("invalid regex accepted")
	}
}

func TestNonHTTPProbeRejectsPrivateTargets(t *testing.T) {
	svc := &UptimeService{}
	for _, protocol := range []string{"tcp", "icmp"} {
		target := "127.0.0.1"
		if protocol == "tcp" {
			target += ":8080"
		}
		if _, err := svc.probeTarget(context.Background(), &models.UptimeMonitor{Protocol: protocol, URL: target}); err == nil {
			t.Fatalf("%s accepted private target", protocol)
		}
	}
}

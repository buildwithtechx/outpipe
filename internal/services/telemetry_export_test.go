package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/services"
)

func TestAnalyticsExporterTinybirdAndClickHouse(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodPost {
			t.Error("analytics must POST")
		}
		if r.URL.Path == "/v0/events" {
			if r.URL.Query().Get("name") != "outpipe_telemetry" || r.Header.Get("Authorization") != "Bearer test-token" {
				t.Error("Tinybird request is missing scoped destination or auth")
			}
		} else {
			user, password, ok := r.BasicAuth()
			if !ok || user != "outpipe" || password != "test-password" || r.URL.Query().Get("query") != "INSERT INTO default.outpipe_telemetry FORMAT JSONEachRow" {
				t.Error("ClickHouse request is missing destination or auth")
			}
		}
		var row struct {
			OrganizationID string `json:"organization_id"`
			Signal         string `json:"signal"`
			Payload        string `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&row); err != nil {
			t.Error(err)
		}
		if row.OrganizationID != "org-a" || row.Signal != "metrics" || !json.Valid([]byte(row.Payload)) {
			t.Error("analytics envelope lost tenant or data")
		}
		w.WriteHeader(http.StatusAccepted)
		if r.URL.Path == "/v0/events" {
			if _, err := w.Write([]byte(`{"successful_rows":1,"quarantined_rows":0}`)); err != nil {
				t.Error(err)
			}
		}
	}))
	defer server.Close()
	exporter, err := services.NewHTTPAnalyticsExporter(config.AnalyticsConfig{TinybirdURL: server.URL, TinybirdToken: "test-token", ClickHouseURL: server.URL, ClickHouseUser: "outpipe", ClickHousePassword: "test-password"})
	if err != nil {
		t.Fatal(err)
	}
	if err := exporter.Export(context.Background(), "metrics", []models.TelemetryMetric{{ID: "metric-a", OrganizationID: "org-a", Name: "latency"}}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected both exporters, got %d", calls.Load())
	}
}

func TestAnalyticsExporterRejectsSQLInjectionAndRedirects(t *testing.T) {
	if _, err := services.NewHTTPAnalyticsExporter(config.AnalyticsConfig{ClickHouseTable: "events; DROP TABLE users"}); err == nil {
		t.Fatal("unsafe table accepted")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:1/credentials", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	exporter, err := services.NewHTTPAnalyticsExporter(config.AnalyticsConfig{TinybirdURL: server.URL, TinybirdToken: "test-token"})
	if err != nil {
		t.Fatal(err)
	}
	if err := exporter.Export(context.Background(), "logs", []models.TelemetryLog{{ID: "a", OrganizationID: "org-a"}}); err == nil {
		t.Fatal("analytics redirect accepted")
	}
}

package services_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func setupTestObservabilityDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.RequestCapture{},
		&models.TelemetrySpan{},
		&models.TelemetryLog{},
	); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}
	return db, func() { _ = sqlDB.Close() }
}

func TestObservabilityServiceLifecycle(t *testing.T) {
	db, cleanup := setupTestObservabilityDB(t)
	defer cleanup()

	repo := repositories.NewGormObservabilityRepository(db)
	svc := services.NewObservabilityService(repo)
	ctx := context.Background()

	orgID := "org-test-obs-1"

	// 1. Ingest Request Capture
	capture := &models.RequestCapture{
		ID:               "cap-101",
		OrganizationID:   orgID,
		TunnelID:         "tun-202",
		Method:           "POST",
		Path:             "/v1/charge",
		StatusCode:       200,
		DurationMs:       45,
		RequestHeaders:   `{"Content-Type":"application/json"}`,
		RequestBody:      `{"amount":5000}`,
		RequestBodySize:  15,
		ResponseHeaders:  `{"Content-Type":"application/json"}`,
		ResponseBody:     `{"status":"succeeded"}`,
		ResponseBodySize: 22,
	}
	if err := svc.IngestCapture(ctx, capture); err != nil {
		t.Fatalf("failed to ingest capture: %v", err)
	}

	gotCap, err := svc.GetCapture(ctx, orgID, "cap-101")
	if err != nil {
		t.Fatalf("failed to get capture: %v", err)
	}
	if gotCap.Method != "POST" || gotCap.Path != "/v1/charge" {
		t.Errorf("unexpected capture: %+v", gotCap)
	}

	caps, err := svc.ListCaptures(ctx, orgID, "tun-202", 10)
	if err != nil || len(caps) != 1 {
		t.Fatalf("expected 1 capture, got: %d, err: %v", len(caps), err)
	}

	// 2. Ingest OTLP Traces
	otlpTraceJSON := []byte(`{
		"resourceSpans": [{
			"scopeSpans": [{
				"spans": [
					{
						"traceId": "0123456789abcdef0123456789abcdef",
						"spanId": "0123456789abcdef",
						"parentSpanId": "",
						"name": "HTTP GET /api/users",
						"kind": 2,
						"startTimeUnixNano": "1600000000000000000",
						"endTimeUnixNano": "1600000000050000000",
						"status": {"code": 1}
					},
					{
						"traceId": "0123456789abcdef0123456789abcdef",
						"spanId": "fedcba9876543210",
						"parentSpanId": "0123456789abcdef",
						"name": "SELECT * FROM users",
						"kind": 3,
						"startTimeUnixNano": "1600000000010000000",
						"endTimeUnixNano": "1600000000040000000",
						"status": {"code": 1}
					}
				]
			}]
		}]
	}`)
	count, err := svc.IngestOTLPTraces(ctx, orgID, otlpTraceJSON)
	if err != nil || count != 2 {
		t.Fatalf("expected 2 spans ingested, got: %d, err: %v", count, err)
	}

	traces, err := svc.ListTraces(ctx, orgID, 10)
	if err != nil || len(traces) == 0 {
		t.Fatalf("expected traces list, got: %v", err)
	}

	waterfall, err := svc.GetTraceWaterfall(ctx, orgID, "0123456789abcdef0123456789abcdef")
	if err != nil || len(waterfall) != 2 {
		t.Fatalf("expected 2 spans in waterfall, got: %d, err: %v", len(waterfall), err)
	}

	// 3. Ingest OTLP Logs
	otlpLogsJSON := []byte(`{
		"resourceLogs": [{
			"scopeLogs": [{
				"logRecords": [
					{
						"timeUnixNano": "1600000000000000000",
						"severityText": "INFO",
						"body": {"stringValue": "Worker started successfully"},
						"traceId": "0123456789abcdef0123456789abcdef"
					},
					{
						"timeUnixNano": "1600000000001000000",
						"severityText": "ERROR",
						"body": {"stringValue": "Connection reset by peer"}
					}
				]
			}]
		}]
	}`)
	logCount, err := svc.IngestOTLPLogs(ctx, orgID, otlpLogsJSON)
	if err != nil || logCount != 2 {
		t.Fatalf("expected 2 logs ingested, got: %d, err: %v", logCount, err)
	}

	logs, err := svc.ListLogs(ctx, orgID, "Connection", "ERROR", 10)
	if err != nil || len(logs) != 1 {
		t.Fatalf("expected 1 matching error log, got: %d, err: %v", len(logs), err)
	}

	// 4. Observability Stats
	stats, err := svc.GetStats(ctx, orgID, "24h")
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalRequests == 0 {
		t.Errorf("expected >0 total requests, got %d", stats.TotalRequests)
	}

	// 5. Request Replay
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Replay-Test") != "active" {
			http.Error(w, "missing header", http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Server-Resp", "echoed")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("replayed-successfully"))
	}))
	defer testServer.Close()

	replayResp, err := svc.ReplayRequest(ctx, models.ReplayRequestInput{
		URL:    testServer.URL,
		Method: "POST",
		Headers: map[string]string{
			"X-Replay-Test": "active",
		},
		RequestBody: "payload-content",
	})
	if err == nil || replayResp != nil {
		t.Fatal("replay to loopback must be rejected")
	}
}

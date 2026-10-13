package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type retryTestExporter struct {
	calls int
	fail  bool
	sent  int
}

func (e *retryTestExporter) Export(_ context.Context, _ string, _ any) error {
	if e.fail {
		return fmt.Errorf("temporary analytics outage")
	}
	return nil
}

func (e *retryTestExporter) ExportItems(_ context.Context, _ string, items []map[string]json.RawMessage) error {
	e.calls++
	if e.fail {
		return fmt.Errorf("temporary analytics outage")
	}
	e.sent += len(items)
	return nil
}

func TestCaptureRetryPreservesIDAndSinglePersistedRow(t *testing.T) {
	db := reviewTestDatabase(t, &models.RequestCapture{}, &models.TelemetryExport{})
	svc := NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	svc.SetExportQueue(repositories.NewGormTelemetryExportRepository(db))
	exporter := &retryTestExporter{fail: true}
	svc.SetTelemetryExporter(exporter)
	capture := &models.RequestCapture{OrganizationID: "org", TunnelID: "tunnel", Method: "GET", Path: "/", StatusCode: 200}
	if err := svc.IngestCapture(context.Background(), capture); err != nil {
		t.Fatalf("ingest must succeed while analytics is down: %v", err)
	}
	created := capture.CreatedAt
	if err := svc.IngestCapture(context.Background(), capture); err != nil {
		t.Fatal(err)
	}
	var rows int64
	if err := db.Model(&models.RequestCapture{}).Count(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if rows != 1 || !capture.CreatedAt.Equal(created) {
		t.Fatal("capture retry duplicated or changed stored event")
	}
	var jobs []models.TelemetryExport
	if err := db.Where("signal = ?", "captures").Find(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].EventID != jobs[1].EventID || jobs[0].EventID != capture.ID {
		t.Fatal("capture retries were not queued with the stable event identity")
	}
}

func TestTelemetryExportQueueRetriesThenDelivers(t *testing.T) {
	db := reviewTestDatabase(t, &models.RequestCapture{}, &models.TelemetryExport{})
	svc := NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	exports := repositories.NewGormTelemetryExportRepository(db)
	svc.SetExportQueue(exports)
	exporter := &retryTestExporter{fail: true}
	svc.SetTelemetryExporter(exporter)
	capture := &models.RequestCapture{OrganizationID: "org", TunnelID: "tunnel", Method: "GET", Path: "/", StatusCode: 200}
	if err := svc.IngestCapture(context.Background(), capture); err != nil {
		t.Fatal(err)
	}
	if err := svc.ProcessPendingExports(context.Background(), 100); err != nil {
		t.Fatal(err)
	}
	if exporter.sent != 0 {
		t.Fatal("failed export was reported as sent")
	}
	queued, err := exports.CountQueued(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if queued != 1 {
		t.Fatalf("failed export left queue depth %d", queued)
	}
	exporter.fail = false
	if err := db.Model(&models.TelemetryExport{}).Where("status = ?", models.TelemetryExportFailed).Update("available_at", time.Now().UTC().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if err := svc.ProcessPendingExports(context.Background(), 100); err != nil {
		t.Fatal(err)
	}
	if exporter.sent != 1 {
		t.Fatalf("recovered export delivered %d rows", exporter.sent)
	}
	queued, err = exports.CountQueued(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if queued != 0 {
		t.Fatalf("delivered export left queue depth %d", queued)
	}
}

func TestTelemetryExportQueueDeadLettersAfterMaxAttempts(t *testing.T) {
	db := reviewTestDatabase(t, &models.RequestCapture{}, &models.TelemetryExport{})
	svc := NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	svc.SetExportQueue(repositories.NewGormTelemetryExportRepository(db))
	svc.SetTelemetryExporter(&retryTestExporter{fail: true})
	capture := &models.RequestCapture{OrganizationID: "org", TunnelID: "tunnel", Method: "GET", Path: "/", StatusCode: 200}
	if err := svc.IngestCapture(context.Background(), capture); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < repositories.TelemetryExportMaxAttempts; i++ {
		if err := db.Model(&models.TelemetryExport{}).Where("status != ?", models.TelemetryExportSent).Update("available_at", time.Now().UTC().Add(-time.Minute)).Error; err != nil {
			t.Fatal(err)
		}
		if err := svc.ProcessPendingExports(context.Background(), 100); err != nil {
			t.Fatal(err)
		}
	}
	var job models.TelemetryExport
	if err := db.Where("event_id = ?", capture.ID).First(&job).Error; err != nil {
		t.Fatal(err)
	}
	if job.Status != models.TelemetryExportFailed || job.Attempts != repositories.TelemetryExportMaxAttempts || job.LastError == "" {
		t.Fatalf("poison export was not dead-lettered: %+v", job)
	}
}

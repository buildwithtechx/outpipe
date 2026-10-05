package services

import (
	"context"
	"fmt"
	"testing"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type retryTestExporter struct {
	calls int
	ids   []string
}

func (e *retryTestExporter) Export(_ context.Context, _ string, data any) error {
	e.calls++
	e.ids = append(e.ids, data.([]models.RequestCapture)[0].ID)
	if e.calls == 1 {
		return fmt.Errorf("temporary analytics outage")
	}
	return nil
}

func TestCaptureRetryPreservesIDAndSinglePersistedRow(t *testing.T) {
	db := reviewTestDatabase(t, &models.RequestCapture{})
	svc := NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	exporter := &retryTestExporter{}
	svc.SetTelemetryExporter(exporter)
	capture := &models.RequestCapture{OrganizationID: "org", TunnelID: "tunnel", Method: "GET", Path: "/", StatusCode: 200}
	if err := svc.IngestCapture(context.Background(), capture); err == nil {
		t.Fatal("export failure not reported")
	}
	created := capture.CreatedAt
	if err := svc.IngestCapture(context.Background(), capture); err != nil {
		t.Fatal(err)
	}
	var rows int64
	if err := db.Model(&models.RequestCapture{}).Count(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if rows != 1 || exporter.ids[0] != exporter.ids[1] || !capture.CreatedAt.Equal(created) {
		t.Fatal("capture retry duplicated or changed stored event")
	}
}

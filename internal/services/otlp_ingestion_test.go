package services_test

import (
	"bytes"
	"context"
	"testing"

	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestOTLPMetricsProtobufAndJSONTenantIsolation(t *testing.T) {
	db, cleanup := setupTestObservabilityDB(t)
	defer cleanup()
	if err := db.AutoMigrate(&models.TelemetryMetric{}); err != nil {
		t.Fatal(err)
	}
	svc := services.NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	request := &collector.ExportMetricsServiceRequest{ResourceMetrics: []*metrics.ResourceMetrics{{ScopeMetrics: []*metrics.ScopeMetrics{{Metrics: []*metrics.Metric{{Name: "requests", Unit: "1", Data: &metrics.Metric_Sum{Sum: &metrics.Sum{IsMonotonic: true, DataPoints: []*metrics.NumberDataPoint{{TimeUnixNano: 1000000000, Value: &metrics.NumberDataPoint_AsInt{AsInt: 7}}}}}}}}}}}}
	wire, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	count, err := svc.IngestMetrics(context.Background(), "org-a", wire, "application/x-protobuf")
	if err != nil || count != 1 {
		t.Fatalf("protobuf metrics count=%d: %v", count, err)
	}
	jsonBody := []byte(`{"resourceMetrics":[{"scopeMetrics":[{"metrics":[{"name":"latency","histogram":{"dataPoints":[{"timeUnixNano":"1000000000","count":"2","sum":25,"bucketCounts":["1","1"],"explicitBounds":[10]}]}}]}]}]}`)
	if _, err := svc.IngestMetrics(context.Background(), "org-a", jsonBody, "application/json"); err != nil {
		t.Fatal(err)
	}
	rows, err := svc.ListMetrics(context.Background(), "org-a", "", 100)
	if err != nil || len(rows) != 2 {
		t.Fatalf("expected 2 stored metrics: %v", err)
	}
	other, err := svc.ListMetrics(context.Background(), "org-b", "", 100)
	if err != nil || len(other) != 0 {
		t.Fatalf("cross-tenant metrics returned: %v", err)
	}
	if _, err := svc.IngestMetrics(context.Background(), "org-a", bytes.Repeat([]byte("x"), services.MaxOTLPBodyBytes+1), "application/json"); err == nil {
		t.Fatal("oversized telemetry accepted")
	}
	if _, err := svc.IngestMetrics(context.Background(), "org-a", []byte{255}, "application/x-protobuf"); err == nil {
		t.Fatal("malformed protobuf accepted")
	}
}

package services

import (
	"bytes"
	"context"
	"strings"
	"testing"

	logsCollector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsCollector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traceCollector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	logs "go.opentelemetry.io/proto/otlp/logs/v1"
	metrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	resource "go.opentelemetry.io/proto/otlp/resource/v1"
	traces "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestOTLPRejectsRecordsBeforeUnmarshalAndExpandedStorage(t *testing.T) {
	spans := make([]*traces.Span, maxOTLPRecords+1)
	for i := range spans {
		spans[i] = &traces.Span{}
	}
	input := &traceCollector.ExportTraceServiceRequest{ResourceSpans: []*traces.ResourceSpans{{ScopeSpans: []*traces.ScopeSpans{{Spans: spans}}}}}
	wire, err := proto.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded := &traceCollector.ExportTraceServiceRequest{}
	if err := decodeOTLP(wire, "application/x-protobuf", decoded); err == nil {
		t.Fatal("protobuf allocation limit bypassed")
	}
	if len(decoded.ResourceSpans) != 0 {
		t.Fatal("oversized repeated fields were decoded before rejection")
	}
	db := reviewTestDatabase(t, &models.TelemetryMetric{})
	svc := NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	metricRows := make([]*metrics.Metric, 20)
	for i := range metricRows {
		metricRows[i] = &metrics.Metric{Name: "requests", Data: &metrics.Metric_Gauge{Gauge: &metrics.Gauge{DataPoints: []*metrics.NumberDataPoint{{TimeUnixNano: 1}}}}}
	}
	request := &metricsCollector.ExportMetricsServiceRequest{ResourceMetrics: []*metrics.ResourceMetrics{{Resource: &resource.Resource{Attributes: []*common.KeyValue{{Key: "large", Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: strings.Repeat("x", 1024*1024)}}}}}, ScopeMetrics: []*metrics.ScopeMetrics{{Metrics: metricRows}}}}}
	wire, err = proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.IngestMetrics(context.Background(), "org", wire, "application/x-protobuf"); err == nil {
		t.Fatal("expanded resource writes were accepted")
	}
	var stored int64
	if err := db.Model(&models.TelemetryMetric{}).Count(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored != 0 {
		t.Fatal("rejected batch was partially persisted")
	}
}

func TestOTLPLogFieldsAndInvalidIdentifiers(t *testing.T) {
	db := reviewTestDatabase(t, &models.TelemetryLog{}, &models.TelemetrySpan{})
	svc := NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	entry := &logs.LogRecord{ObservedTimeUnixNano: 1234567890, Body: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: ""}}, Attributes: []*common.KeyValue{{Key: "component", Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "api"}}}}}
	request := &logsCollector.ExportLogsServiceRequest{ResourceLogs: []*logs.ResourceLogs{{ScopeLogs: []*logs.ScopeLogs{{LogRecords: []*logs.LogRecord{entry}}}}}}
	wire, err := proto.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.IngestLogs(context.Background(), "org", wire, "application/x-protobuf"); err != nil {
		t.Fatal(err)
	}
	stored, err := svc.ListLogs(context.Background(), "org", "", "", 10)
	if err != nil || len(stored) != 1 {
		t.Fatalf("read log: %v", err)
	}
	if stored[0].Body != "" || stored[0].Timestamp.UnixNano() != 1234567890 || !strings.Contains(stored[0].Attributes, "component") || strings.Contains(stored[0].Attributes, "observedTime") {
		t.Fatal("OTLP log fields were not separated correctly")
	}
	span := &traces.Span{TraceId: make([]byte, 16), SpanId: bytes.Repeat([]byte{1}, 8), Name: "test"}
	spanRequest := &traceCollector.ExportTraceServiceRequest{ResourceSpans: []*traces.ResourceSpans{{ScopeSpans: []*traces.ScopeSpans{{Spans: []*traces.Span{span}}}}}}
	wire, err = proto.Marshal(spanRequest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.IngestTraces(context.Background(), "org", wire, "application/x-protobuf"); err == nil {
		t.Fatal("zero trace ID accepted")
	}
	span.TraceId = bytes.Repeat([]byte{1}, 16)
	span.Name = strings.Repeat("x", 256)
	wire, err = proto.Marshal(spanRequest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.IngestTraces(context.Background(), "org", wire, "application/x-protobuf"); err == nil {
		t.Fatal("oversized span name accepted")
	}
}

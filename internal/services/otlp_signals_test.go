package services_test

import (
	"bytes"
	"context"
	"testing"

	logsCollector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	traceCollector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	logs "go.opentelemetry.io/proto/otlp/logs/v1"
	traces "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestOTLPProtobufTracesAndLogs(t *testing.T) {
	db, cleanup := setupTestObservabilityDB(t)
	defer cleanup()
	svc := services.NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	span := &traces.Span{TraceId: bytes.Repeat([]byte{1}, 16), SpanId: bytes.Repeat([]byte{2}, 8), Name: "request", StartTimeUnixNano: 1000000000, EndTimeUnixNano: 1001000000}
	traceRequest := &traceCollector.ExportTraceServiceRequest{ResourceSpans: []*traces.ResourceSpans{{ScopeSpans: []*traces.ScopeSpans{{Spans: []*traces.Span{span}}}}}}
	wire, err := proto.Marshal(traceRequest)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := svc.IngestTraces(context.Background(), "org-a", wire, "application/x-protobuf"); err != nil || count != 1 {
		t.Fatalf("trace count=%d: %v", count, err)
	}
	log := &logs.LogRecord{Body: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "healthy"}}, SeverityText: "INFO"}
	logRequest := &logsCollector.ExportLogsServiceRequest{ResourceLogs: []*logs.ResourceLogs{{ScopeLogs: []*logs.ScopeLogs{{LogRecords: []*logs.LogRecord{log}}}}}}
	wire, err = proto.Marshal(logRequest)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := svc.IngestLogs(context.Background(), "org-a", wire, "application/x-protobuf"); err != nil || count != 1 {
		t.Fatalf("log count=%d: %v", count, err)
	}
	rows, err := svc.ListLogs(context.Background(), "org-b", "", "", 100)
	if err != nil || len(rows) != 0 {
		t.Fatalf("foreign tenant read logs: %v", err)
	}
}

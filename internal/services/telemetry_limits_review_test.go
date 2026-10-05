package services

import (
	"encoding/json"
	"testing"
	"time"

	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
)

func TestOTLPPackedHistogramsRejectBeforeAllocation(t *testing.T) {
	input := &collector.ExportMetricsServiceRequest{ResourceMetrics: []*metrics.ResourceMetrics{{ScopeMetrics: []*metrics.ScopeMetrics{{Metrics: []*metrics.Metric{{Name: "histogram", Data: &metrics.Metric_Histogram{Histogram: &metrics.Histogram{DataPoints: []*metrics.HistogramDataPoint{{BucketCounts: make([]uint64, 300001)}}}}}}}}}}}
	raw, err := proto.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded := &collector.ExportMetricsServiceRequest{}
	if err := decodeOTLP(raw, "application/x-protobuf", decoded); err == nil {
		t.Fatal("packed histogram bypassed allocation limits")
	}
	if len(decoded.ResourceMetrics) != 0 {
		t.Fatal("packed values were allocated before rejection")
	}
}

func TestTelemetryClockBoundsAndSpanExportTime(t *testing.T) {
	future := time.Now().Add(time.Hour)
	if err := validateTelemetryTimestamp(uint64(future.UnixNano())); err == nil {
		t.Fatal("future telemetry timestamp accepted")
	}
	event := time.Now().UTC().Add(-time.Hour)
	created := time.Now().UTC()
	encode := func(value time.Time) json.RawMessage {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	item := map[string]json.RawMessage{"start_time": encode(event), "created_at": encode(created)}
	if !analyticsEventTime(item).Equal(event) {
		t.Fatal("span export used ingestion time")
	}
	item["timestamp"] = encode(future)
	if analyticsEventTime(item).After(created.Add(5 * time.Minute)) {
		t.Fatal("future analytics retention timestamp accepted")
	}
}

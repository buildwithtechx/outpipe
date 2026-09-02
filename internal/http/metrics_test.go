package http

import (
	"testing"

	"outpipe.dev/outpipe/internal/infra/telemetry"
)

func TestMetricsExporterProducesPrometheusCounters(t *testing.T) {
	metrics := telemetry.NewMetricsExporter()
	metrics.IncCounter("outpipe_http_requests_total", 2)
	metrics.IncCounter("outpipe_http_responses_total", 2)
	metrics.IncCounter("outpipe_http_responses_total{status=\"200\"}", 2)

	exported := metrics.ExportPrometheus()
	if !containsMetric(exported, "outpipe_http_requests_total 2") || !containsMetric(exported, "outpipe_http_responses_total 2") || !containsMetric(exported, "outpipe_http_responses_total{status=\"200\"} 2") {
		t.Fatalf("expected HTTP counters in metrics output, got %q", exported)
	}
}

func containsMetric(exported, metric string) bool {
	for _, line := range splitLines(exported) {
		if line == metric {
			return true
		}
	}
	return false
}

func splitLines(value string) []string {
	lines := []string{}
	for len(value) > 0 {
		index := 0
		for index < len(value) && value[index] != '\n' {
			index++
		}
		lines = append(lines, value[:index])
		if index == len(value) {
			break
		}
		value = value[index+1:]
	}
	return lines
}

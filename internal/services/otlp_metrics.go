package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metrics "go.opentelemetry.io/proto/otlp/metrics/v1"
	"outpipe.dev/outpipe/internal/models"
)

func (s *ObservabilityService) IngestMetrics(ctx context.Context, orgID string, raw []byte, contentType string) (int, error) {
	if orgID == "" {
		return 0, fmt.Errorf("organization is required")
	}
	request := &collector.ExportMetricsServiceRequest{}
	if err := decodeOTLP(raw, contentType, request); err != nil {
		return 0, err
	}
	var rows []models.TelemetryMetric
	count := 0
	for _, resource := range request.ResourceMetrics {
		resourceJSON, err := marshalTelemetry(resource.Resource)
		if err != nil {
			return 0, err
		}
		for _, scope := range resource.ScopeMetrics {
			scopeJSON, err := marshalTelemetry(scope.Scope)
			if err != nil {
				return 0, err
			}
			for _, metric := range scope.Metrics {
				kind, points := metricKind(metric)
				if kind == "" || metric.Name == "" {
					return 0, fmt.Errorf("metric name and supported data are required")
				}
				count += points
				if count > maxOTLPRecords || len(rows) >= maxOTLPRecords {
					return 0, fmt.Errorf("OTLP record limit exceeded")
				}
				data, err := marshalTelemetry(metric)
				if err != nil {
					return 0, err
				}
				rows = append(rows, models.TelemetryMetric{ID: uuid.NewString(), OrganizationID: orgID, Name: metric.Name, Unit: metric.Unit, Type: kind, Resource: resourceJSON, Scope: scopeJSON, Data: data, CreatedAt: time.Now().UTC()})
			}
		}
	}
	if err := s.repo.CreateMetrics(ctx, rows); err != nil {
		return 0, &TelemetryUnavailableError{Cause: err}
	}
	if err := s.exportTelemetry(ctx, "metrics", rows); err != nil {
		return 0, err
	}
	return count, nil
}

func metricKind(metric *metrics.Metric) (string, int) {
	switch value := metric.Data.(type) {
	case *metrics.Metric_Gauge:
		return "gauge", len(value.Gauge.DataPoints)
	case *metrics.Metric_Sum:
		return "sum", len(value.Sum.DataPoints)
	case *metrics.Metric_Histogram:
		return "histogram", len(value.Histogram.DataPoints)
	case *metrics.Metric_ExponentialHistogram:
		return "exponential_histogram", len(value.ExponentialHistogram.DataPoints)
	case *metrics.Metric_Summary:
		return "summary", len(value.Summary.DataPoints)
	default:
		return "", 0
	}
}

func (s *ObservabilityService) ListMetrics(ctx context.Context, orgID, name string, limit int) ([]models.TelemetryMetric, error) {
	rows, err := s.repo.ListMetrics(ctx, orgID, name, limit)
	if err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}
	return rows, nil
}

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type ObservabilityService struct {
	repo     repositories.ObservabilityRepository
	exporter TelemetryExporter
}

func NewObservabilityService(repo repositories.ObservabilityRepository) *ObservabilityService {
	return &ObservabilityService{repo: repo}
}

func (s *ObservabilityService) IngestCapture(ctx context.Context, capture *models.RequestCapture) error {
	if capture == nil {
		return fmt.Errorf("capture cannot be nil")
	}
	if capture.OrganizationID == "" || capture.TunnelID == "" {
		return fmt.Errorf("capture organization and tunnel are required")
	}
	if capture.ID == "" {
		capture.ID = uuid.NewString()
	}
	if capture.Timestamp.IsZero() {
		capture.Timestamp = time.Now()
	}
	capture.CreatedAt = time.Now()
	if err := s.repo.CreateRequestCapture(ctx, capture); err != nil {
		return &TelemetryUnavailableError{Cause: err}
	}
	return s.exportTelemetry(ctx, "captures", []models.RequestCapture{*capture})
}

func (s *ObservabilityService) GetStats(ctx context.Context, orgID, timeRange string) (*models.ObservabilityStats, error) {
	stats, err := s.repo.GetStats(ctx, orgID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("fetch observability stats: %w", err)
	}
	return stats, nil
}

func (s *ObservabilityService) ListTraces(ctx context.Context, orgID string, limit int) ([]models.TelemetrySpan, error) {
	traces, err := s.repo.ListTraces(ctx, orgID, limit)
	if err != nil {
		return nil, fmt.Errorf("list traces: %w", err)
	}
	return traces, nil
}

func (s *ObservabilityService) GetTraceWaterfall(ctx context.Context, orgID, traceID string) ([]models.TelemetrySpan, error) {
	spans, err := s.repo.GetTraceSpans(ctx, orgID, traceID)
	if err != nil {
		return nil, fmt.Errorf("get trace waterfall: %w", err)
	}
	return spans, nil
}

func (s *ObservabilityService) ListLogs(ctx context.Context, orgID, search, severity string, limit int) ([]models.TelemetryLog, error) {
	logs, err := s.repo.ListLogs(ctx, orgID, search, severity, limit)
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}
	return logs, nil
}

func (s *ObservabilityService) ListCaptures(ctx context.Context, orgID, tunnelID string, limit int) ([]models.RequestCapture, error) {
	captures, err := s.repo.ListRequestCaptures(ctx, orgID, tunnelID, limit)
	if err != nil {
		return nil, fmt.Errorf("list captures: %w", err)
	}
	return captures, nil
}

func (s *ObservabilityService) GetCapture(ctx context.Context, orgID, captureID string) (*models.RequestCapture, error) {
	capture, err := s.repo.GetRequestCapture(ctx, orgID, captureID)
	if err != nil {
		return nil, fmt.Errorf("get capture: %w", err)
	}
	return capture, nil
}

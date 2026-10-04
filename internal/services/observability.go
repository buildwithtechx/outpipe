package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type ObservabilityService struct {
	repo repositories.ObservabilityRepository
}

func NewObservabilityService(repo repositories.ObservabilityRepository) *ObservabilityService {
	return &ObservabilityService{repo: repo}
}

func (s *ObservabilityService) IngestCapture(ctx context.Context, capture *models.RequestCapture) error {
	if capture == nil {
		return fmt.Errorf("capture cannot be nil")
	}
	if capture.ID == "" {
		capture.ID = generateRandomHex(16)
	}
	if capture.Timestamp.IsZero() {
		capture.Timestamp = time.Now()
	}
	capture.CreatedAt = time.Now()
	if err := s.repo.CreateRequestCapture(ctx, capture); err != nil {
		return fmt.Errorf("ingest request capture: %w", err)
	}
	return nil
}

func (s *ObservabilityService) IngestOTLPTraces(ctx context.Context, orgID string, rawPayload []byte) (int, error) {
	if len(rawPayload) == 0 {
		return 0, fmt.Errorf("empty payload")
	}
	var otlpData struct {
		ResourceSpans []struct {
			ScopeSpans []struct {
				Spans []struct {
					TraceID           string            `json:"traceId"`
					SpanID            string            `json:"spanId"`
					ParentSpanID      string            `json:"parentSpanId"`
					Name              string            `json:"name"`
					Kind              string            `json:"kind"`
					StartTimeUnixNano uint64            `json:"startTimeUnixNano,string"`
					EndTimeUnixNano   uint64            `json:"endTimeUnixNano,string"`
					Attributes        []json.RawMessage `json:"attributes"`
					Status            struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					} `json:"status"`
				} `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	}

	if err := json.Unmarshal(rawPayload, &otlpData); err != nil {
		return 0, fmt.Errorf("unmarshal otlp traces: %w", err)
	}

	var spans []models.TelemetrySpan
	now := time.Now()

	for _, rs := range otlpData.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			for _, sp := range ss.Spans {
				startTime := time.Unix(0, int64(sp.StartTimeUnixNano))
				endTime := time.Unix(0, int64(sp.EndTimeUnixNano))
				durationMs := endTime.Sub(startTime).Milliseconds()
				if durationMs < 0 {
					durationMs = 0
				}
				attrBytes, _ := json.Marshal(sp.Attributes)
				spans = append(spans, models.TelemetrySpan{
					ID:             generateRandomHex(16),
					OrganizationID: orgID,
					TraceID:        sp.TraceID,
					SpanID:         sp.SpanID,
					ParentSpanID:   sp.ParentSpanID,
					Name:           sp.Name,
					Kind:           sp.Kind,
					StartTime:      startTime,
					EndTime:        endTime,
					DurationMs:     durationMs,
					StatusCode:     sp.Status.Code,
					StatusMessage:  sp.Status.Message,
					Attributes:     string(attrBytes),
					CreatedAt:      now,
				})
			}
		}
	}

	if len(spans) == 0 {
		return 0, nil
	}

	if err := s.repo.CreateSpans(ctx, spans); err != nil {
		return 0, fmt.Errorf("store otlp spans: %w", err)
	}

	return len(spans), nil
}

func (s *ObservabilityService) IngestOTLPLogs(ctx context.Context, orgID string, rawPayload []byte) (int, error) {
	if len(rawPayload) == 0 {
		return 0, fmt.Errorf("empty payload")
	}
	var otlpData struct {
		ResourceLogs []struct {
			ScopeLogs []struct {
				LogRecords []struct {
					TimeUnixNano   uint64 `json:"timeUnixNano,string"`
					SeverityText   string `json:"severityText"`
					SeverityNumber int    `json:"severityNumber"`
					Body           struct {
						StringValue string `json:"stringValue"`
					} `json:"body"`
					TraceID    string            `json:"traceId"`
					SpanID     string            `json:"spanId"`
					Attributes []json.RawMessage `json:"attributes"`
				} `json:"logRecords"`
			} `json:"scopeLogs"`
		} `json:"resourceLogs"`
	}

	if err := json.Unmarshal(rawPayload, &otlpData); err != nil {
		return 0, fmt.Errorf("unmarshal otlp logs: %w", err)
	}

	var logs []models.TelemetryLog
	now := time.Now()

	for _, rl := range otlpData.ResourceLogs {
		for _, sl := range rl.ScopeLogs {
			for _, lr := range sl.LogRecords {
				timestamp := time.Unix(0, int64(lr.TimeUnixNano))
				if lr.TimeUnixNano == 0 {
					timestamp = now
				}
				severity := lr.SeverityText
				if severity == "" {
					severity = "INFO"
				}
				body := lr.Body.StringValue
				if body == "" {
					body = "(empty)"
				}
				attrBytes, _ := json.Marshal(lr.Attributes)
				logs = append(logs, models.TelemetryLog{
					ID:             generateRandomHex(16),
					OrganizationID: orgID,
					TraceID:        lr.TraceID,
					SpanID:         lr.SpanID,
					Timestamp:      timestamp,
					Severity:       severity,
					Body:           body,
					Attributes:     string(attrBytes),
					CreatedAt:      now,
				})
			}
		}
	}

	if len(logs) == 0 {
		return 0, nil
	}

	if err := s.repo.CreateLogs(ctx, logs); err != nil {
		return 0, fmt.Errorf("store otlp logs: %w", err)
	}

	return len(logs), nil
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

func generateRandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

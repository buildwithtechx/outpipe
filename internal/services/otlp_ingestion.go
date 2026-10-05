package services

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	logs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	traces "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"outpipe.dev/outpipe/internal/models"
)

func (s *ObservabilityService) IngestOTLPTraces(ctx context.Context, orgID string, raw []byte) (int, error) {
	return s.IngestTraces(ctx, orgID, raw, "application/json")
}

func (s *ObservabilityService) IngestOTLPLogs(ctx context.Context, orgID string, raw []byte) (int, error) {
	return s.IngestLogs(ctx, orgID, raw, "application/json")
}

func (s *ObservabilityService) IngestTraces(ctx context.Context, orgID string, raw []byte, contentType string) (int, error) {
	if orgID == "" {
		return 0, fmt.Errorf("organization is required")
	}
	request := &traces.ExportTraceServiceRequest{}
	if err := decodeOTLP(raw, contentType, request); err != nil {
		return 0, err
	}
	var rows []models.TelemetrySpan
	expandedBytes := 0
	for _, resource := range request.ResourceSpans {
		resourceJSON, err := marshalTelemetry(resource.Resource)
		if err != nil {
			return 0, err
		}
		for _, scope := range resource.ScopeSpans {
			scopeJSON, err := marshalTelemetry(scope.Scope)
			if err != nil {
				return 0, err
			}
			for _, span := range scope.Spans {
				if len(rows) >= maxOTLPRecords {
					return 0, fmt.Errorf("OTLP record limit exceeded")
				}
				if !validTelemetryID(span.TraceId, 16) || !validTelemetryID(span.SpanId, 8) || (len(span.ParentSpanId) != 0 && !validTelemetryID(span.ParentSpanId, 8)) {
					return 0, fmt.Errorf("invalid span identifier")
				}
				if span.EndTimeUnixNano < span.StartTimeUnixNano {
					return 0, fmt.Errorf("span end precedes start")
				}
				if err := validateTelemetryTimestamp(span.StartTimeUnixNano, span.EndTimeUnixNano); err != nil {
					return 0, err
				}
				if len(span.Name) > 255 || len(span.GetStatus().GetMessage()) > 512 {
					return 0, fmt.Errorf("span field exceeds limit")
				}
				attributes, err := marshalTelemetryAttributes(span.Attributes)
				if err != nil {
					return 0, err
				}
				payload, err := marshalTelemetry(span)
				if err != nil {
					return 0, err
				}
				events := make([]json.RawMessage, 0, len(span.Events))
				for _, event := range span.Events {
					data, err := marshalTelemetry(event)
					if err != nil {
						return 0, err
					}
					events = append(events, json.RawMessage(data))
				}
				eventJSON, err := json.Marshal(events)
				if err != nil {
					return 0, fmt.Errorf("encode span events: %w", err)
				}
				if err := reserveTelemetryBytes(&expandedBytes, payload, attributes, string(eventJSON), resourceJSON, scopeJSON, span.Name, span.GetStatus().GetMessage()); err != nil {
					return 0, err
				}
				rows = append(rows, models.TelemetrySpan{ID: uuid.NewString(), OrganizationID: orgID, TraceID: hex.EncodeToString(span.TraceId), SpanID: hex.EncodeToString(span.SpanId), ParentSpanID: hex.EncodeToString(span.ParentSpanId), Name: span.Name, Kind: strings.TrimPrefix(span.Kind.String(), "SPAN_KIND_"), StartTime: telemetryTime(span.StartTimeUnixNano), EndTime: telemetryTime(span.EndTimeUnixNano), DurationMs: int64((span.EndTimeUnixNano - span.StartTimeUnixNano) / 1000000), StatusCode: strings.TrimPrefix(span.GetStatus().GetCode().String(), "STATUS_CODE_"), StatusMessage: span.GetStatus().GetMessage(), Attributes: attributes, CreatedAt: time.Now().UTC()})
				rows[len(rows)-1].Resource = resourceJSON
				rows[len(rows)-1].Scope = scopeJSON
				rows[len(rows)-1].Payload = payload
				rows[len(rows)-1].Events = string(eventJSON)
			}
		}
	}
	if err := s.repo.CreateSpans(ctx, rows); err != nil {
		return 0, &TelemetryUnavailableError{Cause: err}
	}
	if err := s.exportTelemetry(ctx, "spans", rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (s *ObservabilityService) IngestLogs(ctx context.Context, orgID string, raw []byte, contentType string) (int, error) {
	if orgID == "" {
		return 0, fmt.Errorf("organization is required")
	}
	request := &logs.ExportLogsServiceRequest{}
	if err := decodeOTLP(raw, contentType, request); err != nil {
		return 0, err
	}
	var rows []models.TelemetryLog
	expandedBytes := 0
	for _, resource := range request.ResourceLogs {
		resourceJSON, err := marshalTelemetry(resource.Resource)
		if err != nil {
			return 0, err
		}
		for _, scope := range resource.ScopeLogs {
			scopeJSON, err := marshalTelemetry(scope.Scope)
			if err != nil {
				return 0, err
			}
			for _, log := range scope.LogRecords {
				if len(rows) >= maxOTLPRecords {
					return 0, fmt.Errorf("OTLP record limit exceeded")
				}
				if (len(log.TraceId) != 0 && !validTelemetryID(log.TraceId, 16)) || (len(log.SpanId) != 0 && !validTelemetryID(log.SpanId, 8)) {
					return 0, fmt.Errorf("invalid log identifier")
				}
				if len(log.SeverityText) > 32 {
					return 0, fmt.Errorf("log severity exceeds limit")
				}
				if err := validateTelemetryTimestamp(log.TimeUnixNano, log.ObservedTimeUnixNano); err != nil {
					return 0, err
				}
				attributes, err := marshalTelemetryAttributes(log.Attributes)
				if err != nil {
					return 0, err
				}
				timestamp := telemetryTime(log.TimeUnixNano)
				if log.TimeUnixNano == 0 {
					if log.ObservedTimeUnixNano != 0 {
						timestamp = telemetryTime(log.ObservedTimeUnixNano)
					} else {
						timestamp = time.Now().UTC()
					}
				}
				severity := log.SeverityText
				if severity == "" {
					severity = log.SeverityNumber.String()
					if log.SeverityNumber >= 1 && log.SeverityNumber <= 24 {
						severity = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[(int(log.SeverityNumber)-1)/4]
					}
				}
				body, err := marshalTelemetry(log.Body)
				if err != nil {
					return 0, err
				}
				if value, ok := log.GetBody().GetValue().(*common.AnyValue_StringValue); ok {
					body = value.StringValue
				}
				payload, err := marshalTelemetry(log)
				if err != nil {
					return 0, err
				}
				if err := reserveTelemetryBytes(&expandedBytes, payload, attributes, body, resourceJSON, scopeJSON, severity); err != nil {
					return 0, err
				}
				rows = append(rows, models.TelemetryLog{ID: uuid.NewString(), OrganizationID: orgID, TraceID: hex.EncodeToString(log.TraceId), SpanID: hex.EncodeToString(log.SpanId), Timestamp: timestamp, Severity: severity, Body: body, Attributes: attributes, CreatedAt: time.Now().UTC()})
				rows[len(rows)-1].Resource = resourceJSON
				rows[len(rows)-1].Scope = scopeJSON
				rows[len(rows)-1].Payload = payload
			}
		}
	}
	if err := s.repo.CreateLogs(ctx, rows); err != nil {
		return 0, &TelemetryUnavailableError{Cause: err}
	}
	if err := s.exportTelemetry(ctx, "logs", rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}

func marshalTelemetry(message proto.Message) (string, error) {
	if message == nil {
		return "{}", nil
	}
	data, err := protojson.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("encode telemetry: %w", err)
	}
	return string(data), nil
}

func telemetryTime(nanoseconds uint64) time.Time {
	return time.Unix(int64(nanoseconds/1000000000), int64(nanoseconds%1000000000)).UTC()
}

func validTelemetryID(value []byte, length int) bool {
	if len(value) != length {
		return false
	}
	for _, part := range value {
		if part != 0 {
			return true
		}
	}
	return false
}

func marshalTelemetryAttributes(attributes []*common.KeyValue) (string, error) {
	values := make(map[string]json.RawMessage, len(attributes))
	for _, attribute := range attributes {
		value, err := marshalTelemetry(attribute.Value)
		if err != nil {
			return "", err
		}
		values[attribute.Key] = json.RawMessage(value)
	}
	data, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("encode telemetry attributes: %w", err)
	}
	return string(data), nil
}

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func analyticsDestinationName(signal, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(signal)) {
	case "captures":
		return "tunnel_request_captures"
	case "logs":
		return "otel_logs"
	case "spans":
		return "otel_spans"
	case "metrics":
		return "otel_metrics"
	}
	if fallback != "" {
		return fallback
	}
	return "outpipe_telemetry"
}

func buildAnalyticsBody(signal string, items []map[string]json.RawMessage) ([]byte, error) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		payload, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("encode analytics event: %w", err)
		}
		event := make(map[string]any, len(item)+10)
		for k, raw := range item {
			var v any
			if err := json.Unmarshal(raw, &v); err == nil {
				event[k] = v
			} else {
				event[k] = string(raw)
			}
		}
		event["payload"] = string(payload)
		if _, ok := event["signal"]; !ok {
			event["signal"] = signal
		}
		event["timestamp"] = analyticsEventTime(item).Format(time.RFC3339Nano)
		if idVal, ok := event["id"]; ok {
			if _, hasEventID := event["event_id"]; !hasEventID {
				event["event_id"] = idVal
			}
		}
		if _, ok := event["ingested_at"]; !ok {
			event["ingested_at"] = now
		}
		if _, ok := event["retention_days"]; !ok {
			event["retention_days"] = 30
		}
		if sev, ok := event["severity"]; ok {
			if _, hasText := event["severity_text"]; !hasText {
				event["severity_text"] = sev
			}
			if _, hasLevel := event["severity_level"]; !hasLevel {
				event["severity_level"] = strings.ToLower(fmt.Sprint(sev))
			}
		}
		if err := encoder.Encode(event); err != nil {
			return nil, fmt.Errorf("encode analytics row: %w", err)
		}
	}
	return body.Bytes(), nil
}

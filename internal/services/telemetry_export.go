package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type TelemetryExporter interface {
	Export(context.Context, string, any) error
	ExportItems(context.Context, string, []map[string]json.RawMessage) error
}

type HTTPAnalyticsExporter struct {
	config config.AnalyticsConfig
	client *http.Client
}

func NewHTTPAnalyticsExporter(cfg config.AnalyticsConfig) (*HTTPAnalyticsExporter, error) {
	identifier := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,127}$`)
	for _, name := range []string{cfg.ClickHouseDatabase, cfg.ClickHouseTable, cfg.TinybirdDatasource} {
		if name != "" && !identifier.MatchString(name) {
			return nil, fmt.Errorf("invalid analytics table or datasource name")
		}
	}
	for _, endpoint := range []string{cfg.TinybirdURL, cfg.ClickHouseURL} {
		if endpoint == "" {
			continue
		}
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("invalid analytics endpoint")
		}
	}
	if cfg.TinybirdURL != "" && cfg.TinybirdToken == "" {
		return nil, fmt.Errorf("Tinybird token is required")
	}
	if cfg.TinybirdDatasource == "" {
		cfg.TinybirdDatasource = "outpipe_telemetry"
	}
	if cfg.ClickHouseDatabase == "" {
		cfg.ClickHouseDatabase = "default"
	}
	if cfg.ClickHouseTable == "" {
		cfg.ClickHouseTable = "outpipe_telemetry"
	}
	return &HTTPAnalyticsExporter{config: cfg, client: &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func (s *ObservabilityService) SetTelemetryExporter(exporter TelemetryExporter) {
	s.exporter = exporter
}

func (s *ObservabilityService) exportTelemetry(ctx context.Context, signal string, rows any) error {
	if s.exports == nil {
		return &TelemetryUnavailableError{Cause: fmt.Errorf("telemetry export queue is required")}
	}
	data, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode telemetry export batch: %w", err)
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("decode telemetry export batch: %w", err)
	}
	now := time.Now().UTC()
	for _, item := range items {
		var id, organizationID string
		if err := json.Unmarshal(item["id"], &id); err != nil || id == "" {
			return fmt.Errorf("telemetry export row is missing its event identity")
		}
		if err := json.Unmarshal(item["organization_id"], &organizationID); err != nil || organizationID == "" {
			return fmt.Errorf("telemetry export row is missing its organization")
		}
		payload, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("encode telemetry export row: %w", err)
		}
		job := &models.TelemetryExport{OrganizationID: organizationID, Signal: signal, EventID: id, Payload: string(payload), EventTime: analyticsEventTime(item), Status: models.TelemetryExportPending, AvailableAt: now}
		if err := s.exports.Create(ctx, job); err != nil {
			return &TelemetryUnavailableError{Cause: fmt.Errorf("queue %s telemetry export: %w", signal, err)}
		}
	}
	return nil
}

func (s *ObservabilityService) ProcessPendingExports(ctx context.Context, limit int) error {
	if s.exports == nil {
		return fmt.Errorf("telemetry export queue is required")
	}
	if s.exporter == nil {
		return fmt.Errorf("telemetry exporter is required")
	}
	if limit <= 0 {
		limit = 100
	}
	for {
		jobs, err := s.exports.ClaimPending(ctx, time.Now().UTC(), limit)
		if err != nil {
			return err
		}
		if len(jobs) == 0 {
			break
		}
		grouped := map[string][]int{}
		for i := range jobs {
			grouped[jobs[i].Signal] = append(grouped[jobs[i].Signal], i)
		}
		for signal, indexes := range grouped {
			items := make([]map[string]json.RawMessage, 0, len(indexes))
			for _, i := range indexes {
				var item map[string]json.RawMessage
				if err := json.Unmarshal([]byte(jobs[i].Payload), &item); err != nil {
					return fmt.Errorf("decode queued telemetry export: %w", err)
				}
				items = append(items, item)
			}
			if err := s.exporter.ExportItems(ctx, signal, items); err != nil {
				for _, i := range indexes {
					availableAt := time.Now().UTC()
					if jobs[i].Attempts < repositories.TelemetryExportMaxAttempts {
						availableAt = availableAt.Add(webhookRetryDelay(jobs[i].Attempts))
					}
					if markErr := s.exports.MarkFailed(ctx, jobs[i].ID, err.Error(), availableAt); markErr != nil {
						return fmt.Errorf("mark telemetry export %s failed: %w", jobs[i].ID, markErr)
					}
				}
				continue
			}
			now := time.Now().UTC()
			for _, i := range indexes {
				if markErr := s.exports.MarkSent(ctx, jobs[i].ID, now); markErr != nil {
					return fmt.Errorf("mark telemetry export %s sent: %w", jobs[i].ID, markErr)
				}
			}
		}
		if len(jobs) < limit {
			break
		}
	}
	now := time.Now().UTC()
	if _, err := s.exports.PruneExported(ctx, now.AddDate(0, 0, -7), now.AddDate(0, 0, -30)); err != nil {
		return fmt.Errorf("prune telemetry exports: %w", err)
	}
	return nil
}

func (e *HTTPAnalyticsExporter) Export(ctx context.Context, signal string, rows any) error {
	data, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode analytics batch: %w", err)
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("decode analytics batch: %w", err)
	}
	return e.ExportItems(ctx, signal, items)
}

func (e *HTTPAnalyticsExporter) ExportItems(ctx context.Context, signal string, items []map[string]json.RawMessage) error {
	if len(items) == 0 {
		return nil
	}
	if e.config.TinybirdURL == "" && e.config.ClickHouseURL == "" {
		return fmt.Errorf("no analytics export destination configured")
	}
	body, err := buildAnalyticsBody(signal, items)
	if err != nil {
		return err
	}
	if e.config.TinybirdURL != "" {
		ds := analyticsDestinationName(signal, e.config.TinybirdDatasource)
		endpoint := strings.TrimRight(e.config.TinybirdURL, "/") + "/v0/events?" + url.Values{"name": {ds}, "wait": {"true"}}.Encode()
		if err := e.send(ctx, endpoint, body, true, len(items)); err != nil {
			return err
		}
	}
	if e.config.ClickHouseURL != "" {
		table := analyticsDestinationName(signal, e.config.ClickHouseTable)
		query := "INSERT INTO " + e.config.ClickHouseDatabase + "." + table + " FORMAT JSONEachRow"
		endpoint := strings.TrimRight(e.config.ClickHouseURL, "/") + "/?" + url.Values{"query": {query}, "date_time_input_format": {"best_effort"}, "wait_end_of_query": {"1"}}.Encode()
		if err := e.send(ctx, endpoint, body, false, len(items)); err != nil {
			return err
		}
	}
	return nil
}

func analyticsDestinationName(signal, fallback string) string {
	if fallback != "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(signal)) {
	case "captures":
		return "tunnel_request_captures"
	case "logs":
		return "otel_logs"
	case "spans":
		return "otel_spans"
	case "metrics":
		return "otel_metrics"
	default:
		return "outpipe_telemetry"
	}
}

func buildAnalyticsBody(signal string, items []map[string]json.RawMessage) ([]byte, error) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	for _, item := range items {
		payload, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("encode analytics event: %w", err)
		}
		event := make(map[string]any, len(item)+8)
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
		if _, ok := event["timestamp"]; !ok {
			event["timestamp"] = analyticsEventTime(item).Format(time.RFC3339Nano)
		}
		if idVal, ok := event["id"]; ok {
			if _, hasEventID := event["event_id"]; !hasEventID {
				event["event_id"] = idVal
			}
		}
		if err := encoder.Encode(event); err != nil {
			return nil, fmt.Errorf("encode analytics row: %w", err)
		}
	}
	return body.Bytes(), nil
}

func (e *HTTPAnalyticsExporter) send(ctx context.Context, endpoint string, body []byte, tinybird bool, count int) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create analytics request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-ndjson")
	if tinybird {
		request.Header.Set("Authorization", "Bearer "+e.config.TinybirdToken)
	} else {
		request.SetBasicAuth(e.config.ClickHouseUser, e.config.ClickHousePassword)
	}
	response, err := e.client.Do(request)
	if err != nil {
		return fmt.Errorf("analytics request failed")
	}
	defer response.Body.Close()
	ack, err := io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil {
		return fmt.Errorf("read analytics acknowledgement: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("analytics returned HTTP %d", response.StatusCode)
	}
	if len(ack) > 4096 {
		return fmt.Errorf("analytics acknowledgement exceeds limit")
	}
	if tinybird {
		var result struct {
			SuccessfulRows  int `json:"successful_rows"`
			QuarantinedRows int `json:"quarantined_rows"`
		}
		if err := json.Unmarshal(ack, &result); err != nil {
			return fmt.Errorf("decode Tinybird acknowledgement: %w", err)
		}
		if result.QuarantinedRows != 0 || result.SuccessfulRows != count {
			return fmt.Errorf("Tinybird did not accept all telemetry rows")
		}
	} else if response.Header.Get("X-ClickHouse-Exception-Code") != "" || len(bytes.TrimSpace(ack)) != 0 {
		return fmt.Errorf("ClickHouse insert failed")
	}
	return nil
}

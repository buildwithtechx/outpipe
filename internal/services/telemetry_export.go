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
)

type TelemetryExporter interface {
	Export(context.Context, string, any) error
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
	if s.exporter == nil {
		return nil
	}
	if err := s.exporter.Export(ctx, signal, rows); err != nil {
		return &TelemetryUnavailableError{Cause: fmt.Errorf("export %s telemetry: %w", signal, err)}
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
	if len(items) == 0 {
		return nil
	}
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	for _, item := range items {
		payload, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("encode analytics event: %w", err)
		}
		timestamp := analyticsEventTime(item).Format(time.RFC3339Nano)
		event := map[string]any{"id": item["id"], "organization_id": item["organization_id"], "signal": signal, "payload": string(payload), "timestamp": timestamp}
		if err := encoder.Encode(event); err != nil {
			return fmt.Errorf("encode analytics row: %w", err)
		}
	}
	if e.config.TinybirdURL != "" {
		endpoint := strings.TrimRight(e.config.TinybirdURL, "/") + "/v0/events?" + url.Values{"name": {e.config.TinybirdDatasource}, "wait": {"true"}}.Encode()
		if err := e.send(ctx, endpoint, body.Bytes(), true, len(items)); err != nil {
			return err
		}
	}
	if e.config.ClickHouseURL != "" {
		query := "INSERT INTO " + e.config.ClickHouseDatabase + "." + e.config.ClickHouseTable + " FORMAT JSONEachRow"
		endpoint := strings.TrimRight(e.config.ClickHouseURL, "/") + "/?" + url.Values{"query": {query}, "date_time_input_format": {"best_effort"}, "wait_end_of_query": {"1"}}.Encode()
		if err := e.send(ctx, endpoint, body.Bytes(), false, len(items)); err != nil {
			return err
		}
	}
	return nil
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

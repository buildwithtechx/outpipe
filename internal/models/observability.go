package models

import "time"

type RequestCapture struct {
	ID               string    `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID   string    `gorm:"index;size:64;not null" json:"organization_id"`
	TunnelID         string    `gorm:"index;size:64;not null" json:"tunnel_id"`
	Timestamp        time.Time `gorm:"index;not null" json:"timestamp"`
	Method           string    `gorm:"size:16;not null" json:"method"`
	Path             string    `gorm:"size:2048;not null" json:"path"`
	StatusCode       int       `gorm:"index;not null" json:"status_code"`
	DurationMs       int64     `json:"duration_ms"`
	RequestHeaders   string    `gorm:"type:text" json:"request_headers"`
	RequestBody      string    `gorm:"type:text" json:"request_body"`
	RequestBodySize  int64     `json:"request_body_size"`
	ResponseHeaders  string    `gorm:"type:text" json:"response_headers"`
	ResponseBody     string    `gorm:"type:text" json:"response_body"`
	ResponseBodySize int64     `json:"response_body_size"`
	CreatedAt        time.Time `json:"created_at"`
}

type TelemetrySpan struct {
	Resource       string    `gorm:"type:text" json:"resource"`
	Scope          string    `gorm:"type:text" json:"scope"`
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID string    `gorm:"index;size:64;not null" json:"organization_id"`
	TraceID        string    `gorm:"index;size:64;not null" json:"trace_id"`
	SpanID         string    `gorm:"index;size:64;not null" json:"span_id"`
	ParentSpanID   string    `gorm:"size:64" json:"parent_span_id,omitempty"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Kind           string    `gorm:"size:32" json:"kind"`
	StartTime      time.Time `gorm:"index;not null" json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	DurationMs     int64     `json:"duration_ms"`
	StatusCode     string    `gorm:"size:32" json:"status_code"`
	StatusMessage  string    `gorm:"size:512" json:"status_message,omitempty"`
	Attributes     string    `gorm:"type:text" json:"attributes"`
	Events         string    `gorm:"type:text" json:"events"`
	CreatedAt      time.Time `json:"created_at"`
}

type TelemetryLog struct {
	Resource       string    `gorm:"type:text" json:"resource"`
	Scope          string    `gorm:"type:text" json:"scope"`
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID string    `gorm:"index;size:64;not null" json:"organization_id"`
	TraceID        string    `gorm:"index;size:64" json:"trace_id,omitempty"`
	SpanID         string    `gorm:"size:64" json:"span_id,omitempty"`
	Timestamp      time.Time `gorm:"index;not null" json:"timestamp"`
	Severity       string    `gorm:"index;size:32;not null" json:"severity"`
	Body           string    `gorm:"type:text;not null" json:"body"`
	Attributes     string    `gorm:"type:text" json:"attributes"`
	CreatedAt      time.Time `json:"created_at"`
}

type TimeSeriesDataPoint struct {
	Time     string `json:"time"`
	Requests int64  `json:"requests"`
	Errors   int64  `json:"errors"`
	Bytes    int64  `json:"bytes"`
}

type ObservabilityStats struct {
	TotalRequests      int64                 `json:"total_requests"`
	RequestsChange     float64               `json:"requests_change"`
	TotalBytes         int64                 `json:"total_bytes"`
	DataTransferChange float64               `json:"data_transfer_change"`
	P50LatencyMs       int64                 `json:"p50_latency_ms"`
	P95LatencyMs       int64                 `json:"p95_latency_ms"`
	P99LatencyMs       int64                 `json:"p99_latency_ms"`
	ChartData          []TimeSeriesDataPoint `json:"chart_data"`
}

type ReplayRequestInput struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	RequestBody string            `json:"request_body"`
}

type ReplayResponseOutput struct {
	StatusCode int               `json:"status_code"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	DurationMs int64             `json:"duration_ms"`
}

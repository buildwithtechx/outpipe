package repositories

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

type ObservabilityRepository interface {
	CreateMetrics(context.Context, []models.TelemetryMetric) error
	ListMetrics(context.Context, string, string, int) ([]models.TelemetryMetric, error)
	CreateRequestCapture(ctx context.Context, capture *models.RequestCapture) error
	GetRequestCapture(ctx context.Context, orgID, captureID string) (*models.RequestCapture, error)
	ListRequestCaptures(ctx context.Context, orgID, tunnelID string, limit int) ([]models.RequestCapture, error)

	CreateSpans(ctx context.Context, spans []models.TelemetrySpan) error
	ListTraces(ctx context.Context, orgID string, limit int) ([]models.TelemetrySpan, error)
	GetTraceSpans(ctx context.Context, orgID, traceID string) ([]models.TelemetrySpan, error)

	CreateLogs(ctx context.Context, logs []models.TelemetryLog) error
	ListLogs(ctx context.Context, orgID, search, severity string, limit int) ([]models.TelemetryLog, error)

	GetStats(ctx context.Context, orgID string, timeRange string) (*models.ObservabilityStats, error)
}

type GormObservabilityRepository struct {
	db *gorm.DB
}

func NewGormObservabilityRepository(db *gorm.DB) *GormObservabilityRepository {
	return &GormObservabilityRepository{db: db}
}

func (r *GormObservabilityRepository) CreateRequestCapture(ctx context.Context, capture *models.RequestCapture) error {
	if capture == nil {
		return fmt.Errorf("capture is nil")
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(capture).Error; err != nil {
		return fmt.Errorf("create request capture: %w", err)
	}
	var stored models.RequestCapture
	if err := r.db.WithContext(ctx).Where("id = ? AND organization_id = ? AND tunnel_id = ?", capture.ID, capture.OrganizationID, capture.TunnelID).First(&stored).Error; err != nil {
		return fmt.Errorf("read persisted capture: %w", err)
	}
	*capture = stored
	return nil
}

func (r *GormObservabilityRepository) GetRequestCapture(ctx context.Context, orgID, captureID string) (*models.RequestCapture, error) {
	var capture models.RequestCapture
	if err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", captureID, orgID).
		First(&capture).Error; err != nil {
		return nil, fmt.Errorf("get request capture %q: %w", captureID, err)
	}
	return &capture, nil
}

func (r *GormObservabilityRepository) ListRequestCaptures(ctx context.Context, orgID, tunnelID string, limit int) ([]models.RequestCapture, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	if tunnelID != "" {
		query = query.Where("tunnel_id = ?", tunnelID)
	}
	var captures []models.RequestCapture
	if err := query.Order("timestamp desc").Limit(limit).Find(&captures).Error; err != nil {
		return nil, fmt.Errorf("list request captures: %w", err)
	}
	return captures, nil
}

func (r *GormObservabilityRepository) CreateSpans(ctx context.Context, spans []models.TelemetrySpan) error {
	if len(spans) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&spans).Error; err != nil {
		return fmt.Errorf("create telemetry spans: %w", err)
	}
	return nil
}

func (r *GormObservabilityRepository) ListTraces(ctx context.Context, orgID string, limit int) ([]models.TelemetrySpan, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rootSpans []models.TelemetrySpan
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Where("parent_span_id = '' OR parent_span_id IS NULL").
		Order("start_time desc").
		Limit(limit).
		Find(&rootSpans).Error
	if err != nil {
		return nil, fmt.Errorf("list traces: %w", err)
	}
	if len(rootSpans) == 0 {
		err = r.db.WithContext(ctx).
			Where("organization_id = ?", orgID).
			Order("start_time desc").
			Limit(limit).
			Find(&rootSpans).Error
		if err != nil {
			return nil, fmt.Errorf("list fallback traces: %w", err)
		}
	}
	return rootSpans, nil
}

func (r *GormObservabilityRepository) GetTraceSpans(ctx context.Context, orgID, traceID string) ([]models.TelemetrySpan, error) {
	var spans []models.TelemetrySpan
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND trace_id = ?", orgID, traceID).
		Order("start_time asc").
		Find(&spans).Error
	if err != nil {
		return nil, fmt.Errorf("get trace spans for %q: %w", traceID, err)
	}
	return spans, nil
}

func (r *GormObservabilityRepository) CreateLogs(ctx context.Context, logs []models.TelemetryLog) error {
	if len(logs) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&logs).Error; err != nil {
		return fmt.Errorf("create telemetry logs: %w", err)
	}
	return nil
}

func (r *GormObservabilityRepository) ListLogs(ctx context.Context, orgID, search, severity string, limit int) ([]models.TelemetryLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if search != "" {
		query = query.Where("body LIKE ?", "%"+search+"%")
	}
	var logs []models.TelemetryLog
	if err := query.Order("timestamp desc").Limit(limit).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("list telemetry logs: %w", err)
	}
	return logs, nil
}

func (r *GormObservabilityRepository) GetStats(ctx context.Context, orgID string, timeRange string) (*models.ObservabilityStats, error) {
	duration := 24 * time.Hour
	switch timeRange {
	case "1h":
		duration = time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	}
	since := time.Now().Add(-duration)

	var captures []models.RequestCapture
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND timestamp >= ?", orgID, since).
		Order("timestamp asc").
		Find(&captures).Error; err != nil {
		return nil, fmt.Errorf("query captures for stats: %w", err)
	}

	var spans []models.TelemetrySpan
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND start_time >= ?", orgID, since).
		Find(&spans).Error; err != nil {
		return nil, fmt.Errorf("query spans for stats: %w", err)
	}

	totalRequests := int64(len(captures) + len(spans))
	var totalBytes int64
	var durations []int64

	for _, c := range captures {
		totalBytes += c.RequestBodySize + c.ResponseBodySize
		if c.DurationMs > 0 {
			durations = append(durations, c.DurationMs)
		}
	}
	for _, s := range spans {
		if s.DurationMs > 0 {
			durations = append(durations, s.DurationMs)
		}
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	var p50, p95, p99 int64
	if n := len(durations); n > 0 {
		p50 = durations[n*50/100]
		p95 = durations[n*95/100]
		p99 = durations[n*99/100]
	}

	bucketCount := 12
	bucketDuration := duration / time.Duration(bucketCount)
	chartData := make([]models.TimeSeriesDataPoint, bucketCount)

	for i := 0; i < bucketCount; i++ {
		bStart := since.Add(time.Duration(i) * bucketDuration)
		bEnd := bStart.Add(bucketDuration)
		chartData[i] = models.TimeSeriesDataPoint{
			Time: bStart.Format("15:04"),
		}
		for _, c := range captures {
			if c.Timestamp.After(bStart) && c.Timestamp.Before(bEnd) {
				chartData[i].Requests++
				chartData[i].Bytes += c.RequestBodySize + c.ResponseBodySize
				if c.StatusCode >= 400 {
					chartData[i].Errors++
				}
			}
		}
	}

	return &models.ObservabilityStats{
		TotalRequests:      totalRequests,
		RequestsChange:     5.2,
		TotalBytes:         totalBytes,
		DataTransferChange: 3.8,
		P50LatencyMs:       p50,
		P95LatencyMs:       p95,
		P99LatencyMs:       p99,
		ChartData:          chartData,
	}, nil
}

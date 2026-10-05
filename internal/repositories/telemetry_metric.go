package repositories

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
)

func (r *GormObservabilityRepository) CreateMetrics(ctx context.Context, metrics []models.TelemetryMetric) error {
	if len(metrics) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(metrics, 100).Error; err != nil {
		return fmt.Errorf("store telemetry metrics: %w", err)
	}
	return nil
}

func (r *GormObservabilityRepository) ListMetrics(ctx context.Context, orgID, name string, limit int) ([]models.TelemetryMetric, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	if name != "" {
		query = query.Where("name = ?", name)
	}
	var metrics []models.TelemetryMetric
	if err := query.Order("created_at desc").Limit(limit).Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("list telemetry metrics: %w", err)
	}
	return metrics, nil
}

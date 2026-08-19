package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

type UsageRepository interface {
	CreateEvent(context.Context, *models.UsageEvent) error
	UpsertSnapshot(context.Context, *models.UsageSnapshot) error
	FindSnapshot(context.Context, string, time.Time) (models.UsageSnapshot, error)
	ListEvents(context.Context, string, time.Time, time.Time) ([]models.UsageEvent, error)
	ListRequestEvents(context.Context, string, time.Time, time.Time, int) ([]models.UsageEvent, error)
	AggregatePeriod(context.Context, string, time.Time, time.Time) (models.UsageSnapshot, error)
	DeleteBefore(context.Context, string, time.Time) (int64, error)
}

type GormUsageRepository struct{ db *gorm.DB }

func NewUsageRepository(db *gorm.DB) (*GormUsageRepository, error) {

	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GormUsageRepository{db: db}, nil
}

func (r *GormUsageRepository) CreateEvent(ctx context.Context, event *models.UsageEvent) error {

	if event == nil {
		return fmt.Errorf("usage event is required")
	}

	return wrap(r.db.WithContext(ctx).Create(event).Error, "create usage event")
}

func (r *GormUsageRepository) UpsertSnapshot(ctx context.Context, snapshot *models.UsageSnapshot) error {

	if snapshot == nil {
		return fmt.Errorf("usage snapshot is required")
	}

	return wrap(r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organization_id"}, {Name: "period_start"}}, DoUpdates: clause.AssignmentColumns([]string{"period_end", "tunnel_count", "active_connections", "bandwidth_bytes", "request_count", "error_count", "updated_at"})}).Create(snapshot).Error, "upsert usage snapshot")
}

func (r *GormUsageRepository) FindSnapshot(ctx context.Context, organizationID string, periodStart time.Time) (models.UsageSnapshot, error) {
	var snapshot models.UsageSnapshot

	if err := r.db.WithContext(ctx).Where("organization_id = ? AND period_start = ?", organizationID, periodStart).First(&snapshot).Error; err != nil {
		return models.UsageSnapshot{}, mapError(err)
	}

	return snapshot, nil
}

func (r *GormUsageRepository) ListEvents(ctx context.Context, organizationID string, from, to time.Time) ([]models.UsageEvent, error) {
	var events []models.UsageEvent
	err := r.db.WithContext(ctx).Where("organization_id = ? AND occurred_at >= ? AND occurred_at < ?", organizationID, from, to).Order("occurred_at ASC").Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("list usage events: %w", err)
	}

	return events, nil
}

func (r *GormUsageRepository) ListRequestEvents(ctx context.Context, organizationID string, from, to time.Time, limit int) ([]models.UsageEvent, error) {

	if limit <= 0 || limit > 1000 {
		return nil, fmt.Errorf("request log limit must be between 1 and 1000")
	}

	var events []models.UsageEvent
	err := r.db.WithContext(ctx).Where("organization_id = ? AND occurred_at >= ? AND occurred_at < ? AND method <> ''", organizationID, from, to).Order("occurred_at DESC").Limit(limit).Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("list request logs: %w", err)
	}

	return events, nil
}

func (r *GormUsageRepository) AggregatePeriod(ctx context.Context, organizationID string, from, to time.Time) (models.UsageSnapshot, error) {
	var aggregate struct {
		TunnelCount       int
		ActiveConnections int
		BandwidthBytes    int64
		RequestCount      int64
		ErrorCount        int64
	}
	err := r.db.WithContext(ctx).Model(&models.UsageEvent{}).Select("COUNT(DISTINCT tunnel_id) AS tunnel_count, COALESCE(SUM(connections), 0) AS active_connections, COALESCE(SUM(bytes), 0) AS bandwidth_bytes, COALESCE(SUM(CASE WHEN event_type = 'request' THEN 1 ELSE 0 END), 0) AS request_count, COALESCE(SUM(CASE WHEN event_type = 'error' OR status_code >= 400 THEN 1 ELSE 0 END), 0) AS error_count").Where("organization_id = ? AND occurred_at >= ? AND occurred_at < ?", organizationID, from, to).Scan(&aggregate).Error

	if err != nil {
		return models.UsageSnapshot{}, fmt.Errorf("aggregate usage period: %w", err)
	}

	return models.UsageSnapshot{OrganizationID: organizationID, PeriodStart: from, PeriodEnd: to, TunnelCount: aggregate.TunnelCount, ActiveConnections: aggregate.ActiveConnections, BandwidthBytes: aggregate.BandwidthBytes, RequestCount: aggregate.RequestCount, ErrorCount: aggregate.ErrorCount}, nil
}

func (r *GormUsageRepository) DeleteBefore(ctx context.Context, organizationID string, before time.Time) (int64, error) {
	events := r.db.WithContext(ctx).Where("organization_id = ? AND occurred_at < ?", organizationID, before).Delete(&models.UsageEvent{})

	if events.Error != nil {
		return 0, fmt.Errorf("delete usage events before retention cutoff: %w", events.Error)
	}

	snapshots := r.db.WithContext(ctx).Where("organization_id = ? AND period_end < ?", organizationID, before).Delete(&models.UsageSnapshot{})

	if snapshots.Error != nil {
		return events.RowsAffected, fmt.Errorf("delete usage snapshots before retention cutoff: %w", snapshots.Error)
	}

	return events.RowsAffected + snapshots.RowsAffected, nil
}

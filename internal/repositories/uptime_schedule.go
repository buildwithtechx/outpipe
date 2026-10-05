package repositories

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

func (r *GormUptimeRepository) ClaimDueMonitors(ctx context.Context, now time.Time, limit int) ([]models.UptimeMonitor, error) {
	if limit < 1 || limit > 16 {
		limit = 16
	}
	var candidates []models.UptimeMonitor
	due := "next_probe_at <= ? OR (next_probe_at IS NULL AND (last_check_at IS NULL OR last_check_at + interval_seconds * INTERVAL '1 second' <= ?))"
	if r.db.Dialector.Name() == "sqlite" {
		due = "next_probe_at <= ? OR (next_probe_at IS NULL AND (last_check_at IS NULL OR julianday(last_check_at) + interval_seconds / 86400.0 <= julianday(?)))"
	}
	query := r.db.WithContext(ctx).Where("status <> ?", models.MonitorStatusPaused).Where(due, now, now)
	if err := query.Order("next_probe_at asc, created_at asc").Limit(limit).Find(&candidates).Error; err != nil {
		return nil, fmt.Errorf("list due uptime monitors: %w", err)
	}
	claimed := make([]models.UptimeMonitor, 0, len(candidates))
	lease := now.Add(90 * time.Second)
	for _, monitor := range candidates {
		result := r.db.WithContext(ctx).Model(&models.UptimeMonitor{}).Where("id = ? AND status <> ?", monitor.ID, models.MonitorStatusPaused).Where(due, now, now).Update("next_probe_at", lease)
		if result.Error != nil {
			return nil, fmt.Errorf("claim uptime monitor: %w", result.Error)
		}
		if result.RowsAffected == 1 {
			monitor.NextProbeAt = &lease
			claimed = append(claimed, monitor)
		}
	}
	return claimed, nil
}

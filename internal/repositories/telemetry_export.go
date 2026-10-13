package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

const TelemetryExportMaxAttempts = 5

type TelemetryExportRepository interface {
	Create(context.Context, *models.TelemetryExport) error
	ClaimPending(context.Context, time.Time, int) ([]models.TelemetryExport, error)
	MarkSent(context.Context, string, time.Time) error
	MarkFailed(context.Context, string, string, time.Time) error
	PruneExported(context.Context, time.Time, time.Time) (int64, error)
	CountQueued(context.Context) (int64, error)
}

type GormTelemetryExportRepository struct{ db *gorm.DB }

func NewGormTelemetryExportRepository(db *gorm.DB) *GormTelemetryExportRepository {
	return &GormTelemetryExportRepository{db: db}
}

func (r *GormTelemetryExportRepository) Create(ctx context.Context, job *models.TelemetryExport) error {
	if job == nil {
		return fmt.Errorf("telemetry export job is required")
	}
	return wrap(r.db.WithContext(ctx).Create(job).Error, "create telemetry export job")
}

func (r *GormTelemetryExportRepository) ClaimPending(ctx context.Context, now time.Time, limit int) ([]models.TelemetryExport, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("telemetry export limit must be between 1 and 100")
	}
	var jobs []models.TelemetryExport
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		reclaimBefore := now.Add(-10 * time.Minute)
		if err := tx.Where("((status IN ? AND available_at <= ?) OR (status = ? AND updated_at < ?))", []models.TelemetryExportStatus{models.TelemetryExportPending, models.TelemetryExportFailed}, now, models.TelemetryExportSending, reclaimBefore).
			Order("available_at ASC").Limit(limit).Clauses(clauseLockSkipLocked()).Find(&jobs).Error; err != nil {
			return err
		}
		if len(jobs) == 0 {
			return nil
		}
		ids := make([]string, 0, len(jobs))
		for i := range jobs {
			jobs[i].Status = models.TelemetryExportSending
			jobs[i].Attempts++
			ids = append(ids, jobs[i].ID)
		}
		return tx.Model(&models.TelemetryExport{}).Where("id IN ?", ids).Updates(map[string]any{"status": models.TelemetryExportSending, "attempts": gorm.Expr("attempts + 1")}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("claim telemetry exports: %w", err)
	}
	return jobs, nil
}

func (r *GormTelemetryExportRepository) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.TelemetryExport{}).Where("id = ?", id).Updates(map[string]any{"status": models.TelemetryExportSent, "exported_at": sentAt, "last_error": ""}).Error, "mark telemetry export sent")
}

func (r *GormTelemetryExportRepository) MarkFailed(ctx context.Context, id, message string, availableAt time.Time) error {
	return wrap(r.db.WithContext(ctx).Model(&models.TelemetryExport{}).Where("id = ?", id).Updates(map[string]any{"status": models.TelemetryExportFailed, "last_error": message, "available_at": availableAt}).Error, "mark telemetry export failed")
}

func (r *GormTelemetryExportRepository) PruneExported(ctx context.Context, sentBefore, failedBefore time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("(status = ? AND exported_at < ?) OR (status = ? AND attempts >= ? AND updated_at < ?)", models.TelemetryExportSent, sentBefore, models.TelemetryExportFailed, TelemetryExportMaxAttempts, failedBefore).Delete(&models.TelemetryExport{})
	return result.RowsAffected, wrap(result.Error, "prune telemetry exports")
}

func (r *GormTelemetryExportRepository) CountQueued(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.TelemetryExport{}).Where("status IN ?", []models.TelemetryExportStatus{models.TelemetryExportPending, models.TelemetryExportSending, models.TelemetryExportFailed}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count queued telemetry exports: %w", err)
	}
	return count, nil
}

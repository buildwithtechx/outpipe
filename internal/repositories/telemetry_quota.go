package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"outpipe.dev/outpipe/internal/models"
)

var ErrTelemetryQuotaExceeded = errors.New("monthly telemetry allowance exceeded")

const telemetryReceiptCleanupBatch = 1000

type TelemetryQuotaRepository interface {
	AdmitBatch(ctx context.Context, organizationID, month, receiptKey string, bytes int64, limit int64) (bool, error)
	RefundBatch(ctx context.Context, organizationID, month, receiptKey string, bytes int64) error
	UsageForMonth(ctx context.Context, organizationID, month string) (models.OrganizationMonthlyUsage, error)
	DeleteExpiredReceipts(ctx context.Context, cutoff time.Time) (int64, error)
}

type GormTelemetryQuotaRepository struct{ db *gorm.DB }

func NewTelemetryQuotaRepository(db *gorm.DB) (*GormTelemetryQuotaRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormTelemetryQuotaRepository{db: db}, nil
}

func (r *GormTelemetryQuotaRepository) AdmitBatch(ctx context.Context, organizationID, month, receiptKey string, bytes int64, limit int64) (bool, error) {
	if strings.TrimSpace(organizationID) == "" || strings.TrimSpace(month) == "" || strings.TrimSpace(receiptKey) == "" {
		return false, fmt.Errorf("organization, month, and receipt key are required")
	}
	if bytes < 0 {
		return false, fmt.Errorf("telemetry bytes cannot be negative")
	}
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return false, fmt.Errorf("begin telemetry admission: %w", tx.Error)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	usage := models.OrganizationMonthlyUsage{OrganizationID: organizationID, Month: month, Metric: models.TelemetryUsageMetric}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organization_id"}, {Name: "month"}, {Name: "metric"}}, DoNothing: true}).Create(&usage).Error; err != nil {
		return false, fmt.Errorf("ensure telemetry usage counter: %w", err)
	}
	var receipt models.OrganizationTelemetryReceipt
	if err := tx.Where("organization_id = ? AND receipt_key = ?", organizationID, receiptKey).First(&receipt).Error; err == nil {
		if err := tx.Commit().Error; err != nil {
			return false, fmt.Errorf("commit duplicate telemetry admission: %w", err)
		}
		committed = true
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("check telemetry receipt: %w", err)
	}
	var current models.OrganizationMonthlyUsage
	if err := tx.Where("organization_id = ? AND month = ? AND metric = ?", organizationID, month, models.TelemetryUsageMetric).First(&current).Error; err != nil {
		return false, fmt.Errorf("read telemetry usage counter: %w", err)
	}
	if limit >= 0 && current.Used+bytes > limit {
		return false, ErrTelemetryQuotaExceeded
	}
	if limit >= 0 {
		result := tx.Model(&models.OrganizationMonthlyUsage{}).Where("organization_id = ? AND month = ? AND metric = ? AND used + ? <= ?", organizationID, month, models.TelemetryUsageMetric, bytes, limit).Update("used", gorm.Expr("used + ?", bytes))
		if result.Error != nil {
			return false, fmt.Errorf("reserve telemetry quota: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return false, ErrTelemetryQuotaExceeded
		}
	} else if bytes > 0 {
		if err := tx.Model(&models.OrganizationMonthlyUsage{}).Where("organization_id = ? AND month = ? AND metric = ?", organizationID, month, models.TelemetryUsageMetric).Update("used", gorm.Expr("used + ?", bytes)).Error; err != nil {
			return false, fmt.Errorf("reserve unlimited telemetry quota: %w", err)
		}
	}
	now := time.Now().UTC()
	entry := models.OrganizationTelemetryReceipt{OrganizationID: organizationID, ReceiptKey: receiptKey, Month: month, Bytes: bytes, Status: models.TelemetryReceiptAccepted, CreatedAt: now, UpdatedAt: now}
	if err := tx.Create(&entry).Error; err != nil {
		if isDuplicateReceipt(err) {
			return true, nil
		}
		return false, fmt.Errorf("record telemetry receipt: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return false, fmt.Errorf("commit telemetry admission: %w", err)
	}
	committed = true
	return false, nil
}

func (r *GormTelemetryQuotaRepository) RefundBatch(ctx context.Context, organizationID, month, receiptKey string, bytes int64) error {
	if strings.TrimSpace(organizationID) == "" || strings.TrimSpace(receiptKey) == "" {
		return fmt.Errorf("organization and receipt key are required")
	}
	if bytes < 0 {
		return fmt.Errorf("telemetry bytes cannot be negative")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND receipt_key = ?", organizationID, receiptKey).Delete(&models.OrganizationTelemetryReceipt{}).Error; err != nil {
			return fmt.Errorf("remove telemetry receipt: %w", err)
		}
		if bytes == 0 {
			return nil
		}
		if err := tx.Model(&models.OrganizationMonthlyUsage{}).Where("organization_id = ? AND month = ? AND metric = ?", organizationID, month, models.TelemetryUsageMetric).Update("used", gorm.Expr("CASE WHEN used - ? < 0 THEN 0 ELSE used - ? END", bytes, bytes)).Error; err != nil {
			return fmt.Errorf("refund telemetry quota: %w", err)
		}
		return nil
	})
}

func (r *GormTelemetryQuotaRepository) UsageForMonth(ctx context.Context, organizationID, month string) (models.OrganizationMonthlyUsage, error) {
	var usage models.OrganizationMonthlyUsage
	result := r.db.WithContext(ctx).Where("organization_id = ? AND month = ? AND metric = ?", organizationID, month, models.TelemetryUsageMetric).First(&usage)
	if result.Error != nil {
		return models.OrganizationMonthlyUsage{}, mapError(result.Error)
	}
	return usage, nil
}

func (r *GormTelemetryQuotaRepository) DeleteExpiredReceipts(ctx context.Context, cutoff time.Time) (int64, error) {
	if cutoff.IsZero() {
		return 0, fmt.Errorf("receipt cutoff is required")
	}
	result := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Order("created_at ASC").Limit(telemetryReceiptCleanupBatch).Delete(&models.OrganizationTelemetryReceipt{})
	if result.Error != nil {
		return 0, fmt.Errorf("delete expired telemetry receipts: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func isDuplicateReceipt(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "organization_telemetry_receipts") && (strings.Contains(message, "duplicate") || strings.Contains(message, "unique") || strings.Contains(message, "primary"))
}

package models

import "time"

const (
	TelemetryUsageMetric         = "telemetry_bytes"
	TelemetryReceiptAccepted     = "accepted"
	DefaultTelemetryBytesMonthly = int64(1024 * 1024 * 1024)
)

type OrganizationMonthlyUsage struct {
	OrganizationID string    `json:"organizationId" gorm:"type:uuid;primaryKey"`
	Month          string    `json:"month" gorm:"type:varchar(10);primaryKey"`
	Metric         string    `json:"metric" gorm:"type:varchar(64);primaryKey;default:'telemetry_bytes'"`
	Used           int64     `json:"used" gorm:"not null;default:0;check:organization_monthly_usage_nonnegative,used >= 0"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"not null"`
}

func (OrganizationMonthlyUsage) TableName() string { return "organization_monthly_usage" }

type OrganizationTelemetryReceipt struct {
	OrganizationID string    `json:"organizationId" gorm:"type:uuid;primaryKey"`
	ReceiptKey     string    `json:"receiptKey" gorm:"type:varchar(64);primaryKey"`
	Month          string    `json:"month" gorm:"type:varchar(10);not null"`
	Bytes          int64     `json:"bytes" gorm:"not null;check:organization_telemetry_receipts_nonnegative,bytes >= 0"`
	Status         string    `json:"status" gorm:"type:varchar(16);not null;default:'accepted'"`
	CreatedAt      time.Time `json:"createdAt" gorm:"not null;index:organization_telemetry_receipts_created_at_idx"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"not null"`
}

func (OrganizationTelemetryReceipt) TableName() string { return "organization_telemetry_receipts" }

type TelemetryUsage struct {
	Used      int64     `json:"used"`
	Limit     int64     `json:"limit"`
	Unlimited bool      `json:"unlimited"`
	Month     string    `json:"month"`
	StartsAt  time.Time `json:"startsAt"`
	ResetsAt  time.Time `json:"resetsAt"`
}

func TelemetryUsagePeriod(now time.Time) (string, time.Time, time.Time) {
	now = now.UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start.Format("2006-01-02"), start, start.AddDate(0, 1, 0)
}

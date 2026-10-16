package models

import "time"

type TelemetryExportStatus string

const (
	TelemetryExportPending TelemetryExportStatus = "pending"
	TelemetryExportSending TelemetryExportStatus = "sending"
	TelemetryExportSent    TelemetryExportStatus = "sent"
	TelemetryExportFailed  TelemetryExportStatus = "failed"
)

type TelemetryExport struct {
	Base
	OrganizationID string                `json:"organizationId" gorm:"type:uuid;not null;index"`
	Signal         string                `json:"signal" gorm:"size:32;not null;index"`
	EventID        string                `json:"eventId" gorm:"size:64;not null;index"`
	Payload        string                `json:"-" gorm:"type:text;not null"`
	EventTime      time.Time             `json:"eventTime" gorm:"not null;index"`
	Status         TelemetryExportStatus `json:"status" gorm:"size:16;not null;index"`
	Attempts       int                   `json:"attempts" gorm:"not null;default:0"`
	AvailableAt    time.Time             `json:"availableAt" gorm:"not null;index"`
	LastError      string                `json:"lastError,omitempty" gorm:"type:text"`
	ExportedAt     *time.Time            `json:"exportedAt,omitempty"`
}

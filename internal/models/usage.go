package models

import "time"

type UsageSnapshot struct {
	Base
	OrganizationID    string    `json:"organizationId" gorm:"type:uuid;not null;uniqueIndex:organization_period"`
	PeriodStart       time.Time `json:"periodStart" gorm:"not null;uniqueIndex:organization_period"`
	PeriodEnd         time.Time `json:"periodEnd" gorm:"not null"`
	TunnelCount       int       `json:"tunnelCount" gorm:"not null;default:0"`
	ActiveConnections int       `json:"activeConnections" gorm:"not null;default:0"`
	BandwidthBytes    int64     `json:"bandwidthBytes" gorm:"not null;default:0"`
	RequestCount      int64     `json:"requestCount" gorm:"not null;default:0"`
	ErrorCount        int64     `json:"errorCount" gorm:"not null;default:0"`
}

type UsageEvent struct {
	Base
	OrganizationID string    `json:"organizationId" gorm:"type:uuid;not null;index:usage_org_time,priority:1"`
	TunnelID       *string   `json:"tunnelId,omitempty" gorm:"type:uuid;index"`
	EventType      string    `json:"eventType" gorm:"not null;index"`
	Bytes          int64     `json:"bytes" gorm:"not null;default:0"`
	Connections    int       `json:"connections" gorm:"not null;default:0"`
	Method         string    `json:"method,omitempty" gorm:"type:varchar(16)"`
	Path           string    `json:"path,omitempty" gorm:"type:text"`
	StatusCode     int       `json:"statusCode,omitempty" gorm:"default:0"`
	DurationMillis int64     `json:"durationMillis,omitempty" gorm:"default:0"`
	ResponseBytes  int64     `json:"responseBytes,omitempty" gorm:"default:0"`
	ClientIP       *string   `json:"clientIp,omitempty" gorm:"type:inet"`
	OccurredAt     time.Time `json:"occurredAt" gorm:"not null;index:usage_org_time,priority:2"`
}

package models

import "time"

type TelemetryMetric struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID string    `gorm:"index;size:64;not null" json:"organization_id"`
	Name           string    `gorm:"index;size:255;not null" json:"name"`
	Unit           string    `gorm:"size:64" json:"unit"`
	Type           string    `gorm:"size:32" json:"type"`
	Resource       string    `gorm:"type:text" json:"resource"`
	Scope          string    `gorm:"type:text" json:"scope"`
	Data           string    `gorm:"type:text" json:"data"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

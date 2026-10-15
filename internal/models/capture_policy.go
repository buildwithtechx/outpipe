package models

type CapturePolicy struct {
	Base
	OrganizationID string `json:"organizationId" gorm:"type:uuid;not null;index"`
	TunnelID       string `json:"tunnelId" gorm:"size:64;not null;uniqueIndex"`
	Version        int    `json:"version" gorm:"not null;default:1"`
	Rules          string `json:"rules" gorm:"type:text;not null;default:'[]'"`
}

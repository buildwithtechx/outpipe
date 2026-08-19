package models

import "time"

type Tunnel struct {
	Base
	OrganizationID string         `json:"organizationId" gorm:"type:uuid;not null;index"`
	AgentID        *string        `json:"agentId,omitempty" gorm:"type:uuid;index"`
	Name           string         `json:"name" gorm:"not null"`
	Protocol       TunnelProtocol `json:"protocol" gorm:"type:varchar(10);not null"`
	Status         TunnelStatus   `json:"status" gorm:"type:varchar(20);not null;index"`
	TargetHost     string         `json:"targetHost" gorm:"not null"`
	TargetPort     int            `json:"targetPort" gorm:"not null"`
	PublicHostname string         `json:"publicHostname" gorm:"uniqueIndex;not null"`
	PublicPort     *int           `json:"publicPort,omitempty"`
	AccessPolicy   string         `json:"accessPolicy" gorm:"type:jsonb;not null;default:'{}'"`
	Metadata       string         `json:"metadata" gorm:"type:jsonb;not null;default:'{}'"`
	PasswordHash   string         `json:"-" gorm:"type:text"`
	ExpiresAt      *time.Time     `json:"expiresAt,omitempty" gorm:"index"`
	LastActiveAt   *time.Time     `json:"lastActiveAt,omitempty"`
	RevokedAt      *time.Time     `json:"revokedAt,omitempty"`
}

type TunnelToken struct {
	Base
	TunnelID   string     `json:"tunnelId" gorm:"type:uuid;not null;index"`
	Name       string     `json:"name" gorm:"not null"`
	Prefix     string     `json:"prefix" gorm:"uniqueIndex;not null"`
	TokenHash  string     `json:"-" gorm:"uniqueIndex;not null"`
	Scopes     string     `json:"scopes" gorm:"type:jsonb;not null;default:'[]'"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

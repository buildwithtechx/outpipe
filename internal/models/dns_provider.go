package models

import "time"

type DNSProviderCredential struct {
	Base
	OrganizationID string     `json:"organizationId" gorm:"type:uuid;not null;uniqueIndex:idx_dns_credential_org_provider"`
	Provider       string     `json:"provider" gorm:"size:32;not null;uniqueIndex:idx_dns_credential_org_provider"`
	Ciphertext     string     `json:"-" gorm:"type:text;not null"`
	KeyVersion     int        `json:"keyVersion" gorm:"not null;default:1"`
	Label          string     `json:"label,omitempty" gorm:"size:80"`
	LastFour       string     `json:"lastFour,omitempty" gorm:"size:8"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

package models

import "time"

type DomainStatus string

const (
	DomainStatusPending  DomainStatus = "pending"
	DomainStatusVerified DomainStatus = "verified"
	DomainStatusActive   DomainStatus = "active"
	DomainStatusFailed   DomainStatus = "failed"
	DomainStatusRevoked  DomainStatus = "revoked"
)

type Domain struct {
	Base
	OrganizationID       string       `json:"organizationId" gorm:"type:uuid;not null;index"`
	TunnelID             *string      `json:"tunnelId,omitempty" gorm:"type:uuid;index"`
	Hostname             string       `json:"hostname" gorm:"uniqueIndex;not null"`
	Status               DomainStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	VerificationMethod   string       `json:"verificationMethod" gorm:"type:varchar(20);not null"`
	VerificationToken    string       `json:"-" gorm:"not null"`
	VerifiedAt           *time.Time   `json:"verifiedAt,omitempty"`
	CertificateStatus    string       `json:"certificateStatus" gorm:"type:varchar(20);not null;default:'pending'"`
	CertificateExpiresAt *time.Time   `json:"certificateExpiresAt,omitempty"`
	DNSProvider          string       `json:"dnsProvider,omitempty" gorm:"size:32;not null;default:''"`
	BundleGroup          string       `json:"bundleGroup,omitempty" gorm:"size:64;index;not null;default:''"`
	DelegationTarget     string       `json:"delegationTarget,omitempty" gorm:"size:255;not null;default:''"`
}

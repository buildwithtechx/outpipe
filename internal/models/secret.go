package models

import "time"

type SecretProject struct {
	Base
	OrganizationID string `json:"organizationId" gorm:"type:uuid;not null;index"`
	Slug           string `json:"slug" gorm:"size:120;not null"`
	Name           string `json:"name" gorm:"size:120;not null"`
	Description    string `json:"description" gorm:"type:text"`
	CreatedByID    string `json:"createdById" gorm:"type:uuid"`
}

type SecretEnvironment struct {
	Base
	OrganizationID string `json:"organizationId" gorm:"type:uuid;not null;index"`
	ProjectID      string `json:"projectId" gorm:"type:uuid;not null;index"`
	Slug           string `json:"slug" gorm:"size:64;not null"`
	Name           string `json:"name" gorm:"size:64;not null"`
	CreatedByID    string `json:"createdById" gorm:"type:uuid"`
}

type SecretEntry struct {
	Base
	OrganizationID  string     `json:"organizationId" gorm:"type:uuid;not null;index;uniqueIndex:idx_secret_active_key,where:deleted_at IS NULL"`
	ProjectID       string     `json:"projectId" gorm:"type:uuid;not null;index;uniqueIndex:idx_secret_active_key,where:deleted_at IS NULL"`
	EnvironmentID   string     `json:"environmentId" gorm:"type:uuid;not null;index;uniqueIndex:idx_secret_active_key,where:deleted_at IS NULL"`
	Key             string     `json:"key" gorm:"size:256;not null;index;uniqueIndex:idx_secret_active_key,where:deleted_at IS NULL"`
	Comment         string     `json:"comment" gorm:"type:text"`
	DeletedAt       *time.Time `json:"deletedAt,omitempty" gorm:"index"`
	DeletionBatchID *string    `json:"deletionBatchId,omitempty" gorm:"type:uuid"`
}

type SecretVersion struct {
	Base
	OrganizationID string `json:"organizationId" gorm:"type:uuid;not null;index"`
	EntryID        string `json:"entryId" gorm:"type:uuid;not null;index;uniqueIndex:idx_secret_entry_version"`
	Version        int    `json:"version" gorm:"not null;uniqueIndex:idx_secret_entry_version"`
	Ciphertext     string `json:"ciphertext" gorm:"type:text;not null"`
	IV             string `json:"iv" gorm:"type:text;not null"`
	KeyVersion     int    `json:"keyVersion" gorm:"not null;default:1"`
	CreatedByID    string `json:"createdById" gorm:"type:uuid"`
}

type SecretMachineToken struct {
	Base
	OrganizationID string     `json:"organizationId" gorm:"type:uuid;not null;index"`
	ProjectID      *string    `json:"projectId,omitempty" gorm:"type:uuid;index"`
	EnvironmentID  *string    `json:"environmentId,omitempty" gorm:"type:uuid;index"`
	Name           string     `json:"name" gorm:"size:120;not null"`
	Prefix         string     `json:"prefix" gorm:"uniqueIndex;size:32;not null"`
	TokenHash      string     `json:"-" gorm:"uniqueIndex;type:text;not null"`
	Scopes         string     `json:"scopes" gorm:"type:jsonb;not null;default:'[]'"`
	CreatedByID    string     `json:"createdById" gorm:"type:uuid"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty" gorm:"index"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
}

type SecretAuditEvent struct {
	Base
	OrganizationID string  `json:"organizationId" gorm:"type:uuid;not null;index"`
	ProjectID      *string `json:"projectId,omitempty" gorm:"type:uuid;index"`
	EntryID        *string `json:"entryId,omitempty" gorm:"type:uuid;index"`
	Action         string  `json:"action" gorm:"size:64;not null;index"`
	ActorType      string  `json:"actorType" gorm:"size:32;not null"`
	ActorID        string  `json:"actorId" gorm:"type:uuid;not null"`
	Result         string  `json:"result" gorm:"size:32;not null"`
	IPAddress      string  `json:"ipAddress" gorm:"size:64"`
	UserAgent      string  `json:"userAgent" gorm:"type:text"`
}

type SecretShareLink struct {
	ID               string     `json:"id" gorm:"primaryKey;size:64"`
	OrganizationID   *string    `json:"organizationId,omitempty" gorm:"type:uuid;index"`
	ProjectID        *string    `json:"projectId,omitempty" gorm:"type:uuid"`
	EnvironmentID    *string    `json:"environmentId,omitempty" gorm:"type:uuid"`
	CreatedByID      *string    `json:"createdById,omitempty" gorm:"type:uuid"`
	Ciphertext       string     `json:"ciphertext" gorm:"type:text;not null"`
	IV               string     `json:"iv" gorm:"type:text;not null"`
	KeyVerifier      string     `json:"keyVerifier" gorm:"type:text;not null;index"`
	PasswordSalt     *string    `json:"passwordSalt,omitempty" gorm:"type:text"`
	PasswordVerifier *string    `json:"passwordVerifier,omitempty" gorm:"type:text"`
	ContentFormat    string     `json:"contentFormat" gorm:"size:32;not null"`
	ExpiresAt        time.Time  `json:"expiresAt" gorm:"not null;index"`
	MaxViews         int        `json:"maxViews" gorm:"not null;default:1"`
	Views            int        `json:"views" gorm:"not null;default:0"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	LastRevealedAt   *time.Time `json:"lastRevealedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"not null"`
}

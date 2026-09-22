package models

import "time"

type User struct {
	Base
	Email           string     `json:"email" gorm:"uniqueIndex;not null"`
	Name            string     `json:"name" gorm:"not null"`
	Status          UserStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	DeletedAt       *time.Time `json:"-"`
}

type OAuthIdentity struct {
	Base
	UserID         string     `json:"userId" gorm:"type:uuid;not null;uniqueIndex:oauth_provider_subject"`
	Provider       string     `json:"provider" gorm:"type:varchar(20);not null;uniqueIndex:oauth_provider_subject"`
	Subject        string     `json:"subject" gorm:"not null;uniqueIndex:oauth_provider_subject"`
	Email          string     `json:"email,omitempty"`
	AccessToken    string     `json:"-"`
	RefreshToken   string     `json:"-"`
	TokenExpiresAt *time.Time `json:"-"`
}

type Session struct {
	Base
	UserID     string     `json:"userId" gorm:"type:uuid;not null;index"`
	TokenHash  string     `json:"-" gorm:"uniqueIndex;not null"`
	UserAgent  string     `json:"userAgent,omitempty"`
	IPAddress  string     `json:"ipAddress,omitempty"`
	ExpiresAt  time.Time  `json:"expiresAt" gorm:"not null;index"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

type DeviceLogin struct {
	Base
	UserID        *string    `json:"userId,omitempty" gorm:"type:uuid;index"`
	CodeHash      string     `json:"-" gorm:"uniqueIndex;not null"`
	UserTokenHash string     `json:"-" gorm:"index"`
	UserToken     string     `json:"-"`
	Status        string     `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	ExpiresAt     time.Time  `json:"expiresAt" gorm:"not null;index"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	IPAddress     string     `json:"ipAddress,omitempty"`
}

type APIKey struct {
	Base
	UserID         string     `json:"userId" gorm:"type:uuid;not null;index"`
	OrganizationID *string    `json:"organizationId,omitempty" gorm:"type:uuid;index"`
	Name           string     `json:"name" gorm:"not null"`
	Prefix         string     `json:"prefix" gorm:"uniqueIndex;not null"`
	SecretHash     string     `json:"-" gorm:"uniqueIndex;not null"`
	Scopes         string     `json:"scopes" gorm:"type:jsonb;not null;default:'[]'"`
	Source         string     `json:"source,omitempty" gorm:"type:varchar(40);not null;default:''"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
}

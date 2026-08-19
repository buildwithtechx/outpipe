package models

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

type BillingProvider string

const (
	BillingProviderPolar    BillingProvider = "polar"
	BillingProviderPaystack BillingProvider = "paystack"
)

const (
	BillingIntervalMonth = "month"
	BillingIntervalYear  = "year"
)

type Plan struct {
	Base
	Key             string `json:"key" gorm:"uniqueIndex;not null"`
	Name            string `json:"name" gorm:"not null"`
	Description     string `json:"description,omitempty"`
	PriceMinor      int64  `json:"priceMinor" gorm:"not null;default:0"`
	Currency        string `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`
	BillingInterval string `json:"billingInterval" gorm:"type:varchar(10);not null;default:'month'"`
	MaxTunnels      int    `json:"maxTunnels" gorm:"not null"`
	MaxDomains      int    `json:"maxDomains" gorm:"not null"`
	MaxMembers      int    `json:"maxMembers" gorm:"not null"`
	MaxConnections  int    `json:"maxConnections" gorm:"not null;default:0"`
	BandwidthBytes  int64  `json:"bandwidthBytes" gorm:"not null"`
	RetentionDays   int    `json:"retentionDays" gorm:"not null"`
	Features        string `json:"features" gorm:"type:jsonb;not null;default:'{}'"`
	Active          bool   `json:"active" gorm:"not null;default:true"`
}

type Subscription struct {
	Base
	OrganizationID     string             `json:"organizationId" gorm:"type:uuid;uniqueIndex;not null"`
	PlanID             string             `json:"planId" gorm:"type:uuid;not null;index"`
	Provider           BillingProvider    `json:"provider" gorm:"type:varchar(20);not null"`
	Status             SubscriptionStatus `json:"status" gorm:"type:varchar(20);not null;index"`
	CustomerID         string             `json:"customerId,omitempty"`
	ProviderCustomerID string             `json:"-"`
	ProviderSubID      string             `json:"-" gorm:"uniqueIndex"`
	ProviderProductID  string             `json:"-"`
	ProviderAuthCode   string             `json:"-"`
	BillingInterval    string             `json:"billingInterval" gorm:"type:varchar(10);not null"`
	CurrentPeriodEnd   *time.Time         `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd  bool               `json:"cancelAtPeriodEnd" gorm:"not null;default:false"`
	CanceledAt         *time.Time         `json:"canceledAt,omitempty"`
	TrialEndsAt        *time.Time         `json:"trialEndsAt,omitempty"`
}

type BillingEvent struct {
	Base
	Provider        BillingProvider `json:"provider" gorm:"type:varchar(20);not null;uniqueIndex:provider_event"`
	ProviderEventID string          `json:"providerEventId" gorm:"not null;uniqueIndex:provider_event"`
	OrganizationID  *string         `json:"organizationId,omitempty" gorm:"type:uuid;index"`
	EventType       string          `json:"eventType" gorm:"not null"`
	PayloadHash     string          `json:"payloadHash" gorm:"not null"`
	ProcessedAt     *time.Time      `json:"processedAt,omitempty"`
	FailureReason   string          `json:"failureReason,omitempty"`
}

type Invoice struct {
	Base
	OrganizationID  string          `json:"organizationId" gorm:"type:uuid;not null;index"`
	SubscriptionID  string          `json:"subscriptionId" gorm:"type:uuid;not null;index"`
	Provider        BillingProvider `json:"provider" gorm:"type:varchar(20);not null;uniqueIndex:provider_invoice"`
	ProviderInvoice string          `json:"providerInvoice" gorm:"not null;uniqueIndex:provider_invoice"`
	AmountMinor     int64           `json:"amountMinor" gorm:"not null"`
	Currency        string          `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`
	Status          string          `json:"status" gorm:"type:varchar(20);not null;default:'paid'"`
	InvoiceURL      string          `json:"invoiceUrl,omitempty"`
	PDFURL          string          `json:"pdfUrl,omitempty"`
	PaidAt          *time.Time      `json:"paidAt,omitempty"`
}

type Receipt struct {
	Base
	OrganizationID string    `json:"organizationId" gorm:"type:uuid;not null;index"`
	InvoiceID      string    `json:"invoiceId" gorm:"type:uuid;not null;index"`
	ReceiptNumber  string    `json:"receiptNumber" gorm:"not null;uniqueIndex"`
	AmountMinor    int64     `json:"amountMinor" gorm:"not null"`
	Currency       string    `json:"currency" gorm:"type:varchar(3);not null"`
	IssuedAt       time.Time `json:"issuedAt" gorm:"not null"`
}

type BillingCredential struct {
	Base
	OrganizationID string          `json:"organizationId" gorm:"type:uuid;uniqueIndex:billing_credential_scope;not null"`
	Provider       BillingProvider `json:"provider" gorm:"type:varchar(20);uniqueIndex:billing_credential_scope;not null"`
	Kind           string          `json:"kind" gorm:"type:varchar(40);uniqueIndex:billing_credential_scope;not null"`
	Ciphertext     string          `json:"-" gorm:"not null"`
	RotatedAt      time.Time       `json:"rotatedAt" gorm:"not null"`
	ExpiresAt      *time.Time      `json:"expiresAt,omitempty"`
	RevokedAt      *time.Time      `json:"-"`
}

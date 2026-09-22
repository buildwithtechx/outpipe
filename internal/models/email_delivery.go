package models

import "time"

type EmailDeliveryStatus string

const (
	EmailDeliveryPending EmailDeliveryStatus = "pending"
	EmailDeliverySending EmailDeliveryStatus = "sending"
	EmailDeliverySent    EmailDeliveryStatus = "sent"
	EmailDeliveryFailed  EmailDeliveryStatus = "failed"
)

type EmailDelivery struct {
	Base
	To          string              `json:"to" gorm:"not null;index"`
	Subject     string              `json:"subject" gorm:"not null"`
	HTML        string              `json:"-" gorm:"type:text;not null"`
	Status      EmailDeliveryStatus `json:"status" gorm:"not null;index"`
	Attempts    int                 `json:"attempts" gorm:"not null;default:0"`
	AvailableAt time.Time           `json:"availableAt" gorm:"not null;index"`
	LastError   string              `json:"lastError,omitempty" gorm:"type:text"`
	SentAt      *time.Time          `json:"sentAt,omitempty"`
}

package models

import "time"

type WebhookDeliveryStatus string

const (
	WebhookDeliveryPending   WebhookDeliveryStatus = "pending"
	WebhookDeliverySending   WebhookDeliveryStatus = "sending"
	WebhookDeliveryDelivered WebhookDeliveryStatus = "delivered"
	WebhookDeliveryFailed    WebhookDeliveryStatus = "failed"
)

type WebhookEvent string

const (
	WebhookEventTunnelConnected    WebhookEvent = "tunnel.connected"
	WebhookEventTunnelDisconnected WebhookEvent = "tunnel.disconnected"
	WebhookEventTunnelRevoked      WebhookEvent = "tunnel.revoked"
)

func ValidWebhookEvent(event string) bool {
	return event == string(WebhookEventTunnelConnected) || event == string(WebhookEventTunnelDisconnected) || event == string(WebhookEventTunnelRevoked)
}

type WebhookSubscription struct {
	Base
	OrganizationID  string     `json:"organizationId" gorm:"type:uuid;not null;index"`
	Name            string     `json:"name" gorm:"not null"`
	URL             string     `json:"url" gorm:"not null"`
	Events          string     `json:"events" gorm:"type:jsonb;not null;default:'[]'"`
	Secret          string     `json:"-" gorm:"not null"`
	LastDeliveredAt *time.Time `json:"lastDeliveredAt,omitempty"`
}

type WebhookDelivery struct {
	Base
	SubscriptionID string                `json:"subscriptionId" gorm:"type:uuid;not null;uniqueIndex:sub_event"`
	EventID        string                `json:"eventId" gorm:"not null;uniqueIndex:sub_event"`
	EventType      string                `json:"eventType" gorm:"not null"`
	Payload        string                `json:"-" gorm:"type:text;not null"`
	Status         WebhookDeliveryStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Attempts       int                   `json:"attempts" gorm:"not null;default:0"`
	AvailableAt    time.Time             `json:"-" gorm:"not null;index"`
	Error          string                `json:"-" gorm:"type:text"`
	DeliveredAt    *time.Time            `json:"deliveredAt,omitempty"`
	Subscription   *WebhookSubscription  `json:"-" gorm:"foreignKey:SubscriptionID"`
}

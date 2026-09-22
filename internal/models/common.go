package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

import "time"

type Base struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (b *Base) BeforeCreate(_ *gorm.DB) error {

	if b.ID == "" {
		b.ID = uuid.NewString()
	}

	return nil
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
	MemberRoleViewer MemberRole = "viewer"
)

type TunnelProtocol string

const (
	TunnelProtocolHTTP  TunnelProtocol = "http"
	TunnelProtocolHTTPS TunnelProtocol = "https"
	TunnelProtocolTCP   TunnelProtocol = "tcp"
	TunnelProtocolUDP   TunnelProtocol = "udp"
)

type TunnelStatus string

const (
	TunnelStatusCreated      TunnelStatus = "created"
	TunnelStatusConnecting   TunnelStatus = "connecting"
	TunnelStatusActive       TunnelStatus = "active"
	TunnelStatusDisconnected TunnelStatus = "disconnected"
	TunnelStatusExpired      TunnelStatus = "expired"
	TunnelStatusRevoked      TunnelStatus = "revoked"
)

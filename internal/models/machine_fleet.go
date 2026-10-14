package models

type MachineFleet struct {
	Base
	OrganizationID string `json:"organizationId" gorm:"type:uuid;not null;index"`
	Name           string `json:"name" gorm:"size:120;not null"`
	Description    string `json:"description" gorm:"type:text"`
}

type MachineFleetToken struct {
	Base
	OrganizationID string `json:"organizationId" gorm:"type:uuid;not null;index"`
	FleetID        string `json:"fleetId" gorm:"type:uuid;not null;index"`
	Name           string `json:"name" gorm:"size:120;not null"`
	MachineTokenID string `json:"machineTokenId" gorm:"type:uuid;not null;index"`
	Label          string `json:"label" gorm:"size:120"`
}

type MachineFleetTunnel struct {
	FleetID        string `json:"fleetId" gorm:"type:uuid;primaryKey"`
	TunnelID       string `json:"tunnelId" gorm:"type:uuid;primaryKey"`
	DesiredStatus  string `json:"desiredStatus" gorm:"size:32;not null;default:'active'"`
	TargetHost     string `json:"targetHost" gorm:"size:253;not null"`
	TargetPort     int    `json:"targetPort" gorm:"not null"`
	Protocol       string `json:"protocol" gorm:"size:16;not null"`
	Name           string `json:"name" gorm:"size:120;not null"`
	PublicHostname string `json:"publicHostname" gorm:"size:253;not null"`
}

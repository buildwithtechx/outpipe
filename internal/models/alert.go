package models

import "time"

type AlertState string

const (
	AlertStateOK     AlertState = "ok"
	AlertStateFiring AlertState = "firing"
)

type AlertImpact string

const (
	AlertImpactDegraded AlertImpact = "degraded"
	AlertImpactOutage   AlertImpact = "outage"
)

type ObservabilityAlert struct {
	Base
	OrganizationID       string      `json:"organizationId" gorm:"type:uuid;not null;index"`
	Name                 string      `json:"name" gorm:"size:255;not null"`
	MonitorIDs           string      `json:"monitorIds" gorm:"type:text;not null"`
	ConsecutiveFailures  int         `json:"consecutiveFailures" gorm:"not null;default:3"`
	ConsecutiveRecovered int         `json:"consecutiveRecoveries" gorm:"not null;default:2"`
	EvalIntervalSeconds  int         `json:"evalIntervalSeconds" gorm:"not null;default:60"`
	State                AlertState  `json:"state" gorm:"size:16;not null;default:'ok';index"`
	FailureStreak        int         `json:"failureStreak" gorm:"not null;default:0"`
	RecoveryStreak       int         `json:"recoveryStreak" gorm:"not null;default:0"`
	Enabled              bool        `json:"enabled" gorm:"not null;default:true;index"`
	StatusPageTitle      string      `json:"statusPageTitle,omitempty" gorm:"size:255"`
	StatusPageImpact     AlertImpact `json:"statusPageImpact,omitempty" gorm:"size:16"`
	LastEvaluatedAt      *time.Time  `json:"lastEvaluatedAt,omitempty"`
	LastStateChangedAt   *time.Time  `json:"lastStateChangedAt,omitempty"`
	NextEvaluationAt     time.Time   `json:"nextEvaluationAt" gorm:"not null;index"`
}

type AlertIncident struct {
	Base
	OrganizationID string     `json:"organizationId" gorm:"type:uuid;not null;index"`
	AlertID        string     `json:"alertId" gorm:"type:uuid;not null;uniqueIndex"`
	IncidentID     string     `json:"incidentId" gorm:"size:64;not null"`
	StartedAt      time.Time  `json:"startedAt" gorm:"not null"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
}

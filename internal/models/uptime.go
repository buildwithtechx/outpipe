package models

import (
	"time"
)

type MonitorStatus string

const (
	MonitorStatusUp       MonitorStatus = "up"
	MonitorStatusDown     MonitorStatus = "down"
	MonitorStatusDegraded MonitorStatus = "degraded"
	MonitorStatusPaused   MonitorStatus = "paused"
)

type IncidentStatus string

const (
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusIdentified    IncidentStatus = "identified"
	IncidentStatusMonitoring    IncidentStatus = "monitoring"
	IncidentStatusResolved      IncidentStatus = "resolved"
)

type IncidentSeverity string

const (
	IncidentSeverityMinor    IncidentSeverity = "minor"
	IncidentSeverityMajor    IncidentSeverity = "major"
	IncidentSeverityCritical IncidentSeverity = "critical"
)

type UptimeMonitor struct {
	ID              string        `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID  string        `gorm:"index;not null;size:64" json:"organizationId"`
	Name            string        `gorm:"not null;size:255" json:"name"`
	URL             string        `gorm:"not null;size:1024" json:"url"`
	Protocol        string        `gorm:"not null;size:32;default:'https'" json:"protocol"`
	Method          string        `gorm:"not null;size:16;default:'GET'" json:"method"`
	IntervalSeconds int           `gorm:"not null;default:60" json:"intervalSeconds"`
	TimeoutSeconds  int           `gorm:"not null;default:10" json:"timeoutSeconds"`
	Status          MonitorStatus `gorm:"not null;size:32;default:'up'" json:"status"`
	LatencyMs       int64         `gorm:"default:0" json:"latencyMs"`
	UptimeRatio     float64       `gorm:"default:100.0" json:"uptimeRatio"`
	LastCheckAt     *time.Time    `json:"lastCheckAt,omitempty"`
	CreatedAt       time.Time     `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time     `gorm:"not null" json:"updatedAt"`
}

func (UptimeMonitor) TableName() string {
	return "uptime_monitors"
}

type UptimeCheck struct {
	ID           string    `gorm:"primaryKey;size:64" json:"id"`
	MonitorID    string    `gorm:"index;not null;size:64" json:"monitorId"`
	StatusCode   int       `gorm:"not null" json:"statusCode"`
	LatencyMs    int64     `gorm:"not null" json:"latencyMs"`
	Success      bool      `gorm:"not null" json:"success"`
	ErrorMessage string    `gorm:"size:1024" json:"errorMessage,omitempty"`
	CheckedAt    time.Time `gorm:"index;not null" json:"checkedAt"`
}

func (UptimeCheck) TableName() string {
	return "uptime_checks"
}

type UptimeIncident struct {
	ID             string           `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID string           `gorm:"index;not null;size:64" json:"organizationId"`
	Title          string           `gorm:"not null;size:255" json:"title"`
	Status         IncidentStatus   `gorm:"not null;size:32;default:'investigating'" json:"status"`
	Severity       IncidentSeverity `gorm:"not null;size:32;default:'minor'" json:"severity"`
	StartedAt      time.Time        `gorm:"not null" json:"startedAt"`
	ResolvedAt     *time.Time       `json:"resolvedAt,omitempty"`
	CreatedAt      time.Time        `gorm:"not null" json:"createdAt"`
	UpdatedAt      time.Time        `gorm:"not null" json:"updatedAt"`

	Updates []UptimeIncidentUpdate `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE" json:"updates,omitempty"`
}

func (UptimeIncident) TableName() string {
	return "uptime_incidents"
}

type UptimeIncidentUpdate struct {
	ID         string         `gorm:"primaryKey;size:64" json:"id"`
	IncidentID string         `gorm:"index;not null;size:64" json:"incidentId"`
	Status     IncidentStatus `gorm:"not null;size:32" json:"status"`
	Message    string         `gorm:"not null;type:text" json:"message"`
	CreatedAt  time.Time      `gorm:"not null" json:"createdAt"`
}

func (UptimeIncidentUpdate) TableName() string {
	return "uptime_incident_updates"
}

type UptimeStatusPage struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	OrganizationID string    `gorm:"uniqueIndex;not null;size:64" json:"organizationId"`
	Slug           string    `gorm:"uniqueIndex;not null;size:128" json:"slug"`
	Title          string    `gorm:"not null;size:255" json:"title"`
	Description    string    `gorm:"size:1024" json:"description,omitempty"`
	CustomDomain   string    `gorm:"index;size:255" json:"customDomain,omitempty"`
	Published      bool      `gorm:"not null;default:true" json:"published"`
	CreatedAt      time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"not null" json:"updatedAt"`
}

func (UptimeStatusPage) TableName() string {
	return "uptime_status_pages"
}

type UptimeSubscriber struct {
	ID           string     `gorm:"primaryKey;size:64" json:"id"`
	StatusPageID string     `gorm:"index;not null;size:64" json:"statusPageId"`
	Email        string     `gorm:"not null;size:255" json:"email"`
	Confirmed    bool       `gorm:"not null;default:false" json:"confirmed"`
	ConfirmedAt  *time.Time `json:"confirmedAt,omitempty"`
	CreatedAt    time.Time  `gorm:"not null" json:"createdAt"`
}

func (UptimeSubscriber) TableName() string {
	return "uptime_subscribers"
}

package models

import "time"

type AIAgentThread struct {
	Base
	OrganizationID   string            `json:"organizationId" gorm:"type:uuid;not null;index"`
	UserID           string            `json:"userId" gorm:"type:uuid;not null;index"`
	Title            string            `json:"title" gorm:"type:varchar(255);not null"`
	SourceResourceID string            `json:"sourceResourceId,omitempty" gorm:"type:varchar(255)"`
	Messages         []AIAgentMessage  `json:"messages,omitempty" gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE"`
	CreatedAt        time.Time         `json:"createdAt" gorm:"not null"`
	UpdatedAt        time.Time         `json:"updatedAt" gorm:"not null"`
}

type AIAgentMessage struct {
	Base
	ThreadID     string    `json:"threadId" gorm:"type:uuid;not null;index"`
	Role         string    `json:"role" gorm:"type:varchar(20);not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	Status       string    `json:"status" gorm:"type:varchar(20);not null;default:'complete'"`
	StepsJSON    string    `json:"stepsJson,omitempty" gorm:"type:text"`
	EvidenceJSON string    `json:"evidenceJson,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt" gorm:"not null"`
}

type AIAgentStep struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type AIAgentEvidence struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	Href         string                 `json:"href"`
	ObservedAt   string                 `json:"observedAt"`
	Presentation map[string]interface{} `json:"presentation,omitempty"`
}

type AIChatRequest struct {
	ThreadID         string `json:"threadId,omitempty"`
	Message          string `json:"message"`
	SourceResourceID string `json:"sourceRequestId,omitempty"`
	Provider         string `json:"provider,omitempty"`
	Model            string `json:"model,omitempty"`
}

type AIStreamEvent struct {
	Type     string            `json:"type"`
	RunID    string            `json:"runId,omitempty"`
	Step     *AIAgentStep      `json:"step,omitempty"`
	Evidence []AIAgentEvidence `json:"evidence,omitempty"`
	Delta    string            `json:"delta,omitempty"`
	Status   string            `json:"status,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type AIAgentConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Enabled  bool   `json:"enabled"`
}

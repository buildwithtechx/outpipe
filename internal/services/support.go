package services

import (
	"context"
	"fmt"
	"strings"
)

type SupportMailer interface {
	SendContact(context.Context, string, string, string, string) error
	SendBugReport(context.Context, string, string, string, string, string, string, string) error
}

type SupportService struct{ mailer SupportMailer }

func NewSupportService(mailer SupportMailer) (*SupportService, error) {
	if mailer == nil {
		return nil, fmt.Errorf("support mailer is required")
	}
	return &SupportService{mailer: mailer}, nil
}

func (s *SupportService) Contact(ctx context.Context, name, email, topic, message string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(message) == "" {
		return fmt.Errorf("name, email, and message are required")
	}
	return s.mailer.SendContact(ctx, name, email, topic, message)
}

func (s *SupportService) BugReport(ctx context.Context, name, email, category, summary, reproduction, expected, actual string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" || strings.TrimSpace(summary) == "" || strings.TrimSpace(actual) == "" {
		return fmt.Errorf("name, email, summary, and actual result are required")
	}
	return s.mailer.SendBugReport(ctx, name, email, category, summary, reproduction, expected, actual)
}

package mail

import (
	"bytes"
	"fmt"
	"html/template"

	"outpipe.dev/outpipe/templates"
)

type AccountUpdateData struct {
	Event        string
	DashboardURL string
}

type WelcomeData struct {
	Name         string
	DashboardURL string
}

type OrganizationInviteData struct {
	InviterName      string
	OrganizationName string
	Role             string
	InvitationLink   string
}

type PaymentFailedData struct {
	Name              string
	PlanName          string
	Amount            string
	BillingURL        string
	AttemptsRemaining int
	AttemptsKnown     bool
}

type SubscriptionResetData struct {
	Name             string
	OrganizationName string
	PreviousPlan     string
	DashboardURL     string
}

type BillingUpdateData struct {
	Status       string
	DashboardURL string
}

type templateRenderer struct {
	html *template.Template
}

func newTemplateRenderer() (*templateRenderer, error) {
	html, err := template.ParseFS(templates.Email, "*.tmpl")

	if err != nil {
		return nil, fmt.Errorf("parse mail templates: %w", err)
	}

	return &templateRenderer{html: html}, nil
}

func (r *templateRenderer) render(name string, data any) (string, error) {
	var html bytes.Buffer

	if err := r.html.ExecuteTemplate(&html, name+".tmpl", data); err != nil {
		return "", fmt.Errorf("render mail template: %w", err)
	}

	return html.String(), nil
}

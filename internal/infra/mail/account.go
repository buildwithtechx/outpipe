package mail

import (
	"context"
	"fmt"
)

type AccountMailer struct {
	client    *ZeptoClient
	renderer  *templateRenderer
	dashboard string
}

func NewAccountMailer(client *ZeptoClient, dashboardURL string) (*AccountMailer, error) {

	if client == nil {
		return nil, fmt.Errorf("zepto client is required")
	}

	renderer, err := newTemplateRenderer()

	if err != nil {
		return nil, err
	}

	return &AccountMailer{client: client, renderer: renderer, dashboard: dashboardURL}, nil
}

func (m *AccountMailer) SendAccountUpdate(ctx context.Context, email, event string) error {
	html, err := m.renderer.render("account-update", AccountUpdateData{Event: event, DashboardURL: m.dashboard})

	if err != nil {
		return err
	}

	return m.client.Send(ctx, Message{To: email, Subject: "Outpipe account update", HTML: html})
}

func (m *AccountMailer) SendWelcome(ctx context.Context, email, name string) error {
	html, err := m.renderer.render("welcome", WelcomeData{Name: name, DashboardURL: m.dashboard})

	if err != nil {
		return err
	}

	return m.client.Send(ctx, Message{To: email, Subject: "Welcome to Outpipe", HTML: html})
}

func (m *AccountMailer) SendOrganizationInvite(ctx context.Context, email, inviterName, organizationName, role, invitationLink string) error {
	html, err := m.renderer.render("organization-invite", OrganizationInviteData{InviterName: inviterName, OrganizationName: organizationName, Role: role, InvitationLink: invitationLink})

	if err != nil {
		return err
	}

	subject := "You’re invited to join " + organizationName + " on Outpipe"
	return m.client.Send(ctx, Message{To: email, Subject: subject, HTML: html})
}

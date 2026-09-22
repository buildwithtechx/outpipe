package mail

import (
	"context"
	"fmt"
	"html"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type OutboxMailer struct {
	deliveries repositories.EmailRepository
	renderer   *templateRenderer
	dashboard  string
	resolve    func(context.Context, string) (string, error)
	support    string
}

func NewOutboxMailer(deliveries repositories.EmailRepository, resolve func(context.Context, string) (string, error), dashboardURL, supportAddress string) (*OutboxMailer, error) {
	if deliveries == nil || resolve == nil {
		return nil, fmt.Errorf("email repository is required")
	}
	renderer, err := newTemplateRenderer()
	if err != nil {
		return nil, err
	}
	return &OutboxMailer{deliveries: deliveries, renderer: renderer, dashboard: dashboardURL, resolve: resolve, support: supportAddress}, nil
}

func (m *OutboxMailer) enqueue(ctx context.Context, message Message) error {
	return m.deliveries.Create(ctx, &models.EmailDelivery{To: message.To, Subject: message.Subject, HTML: message.HTML, Status: models.EmailDeliveryPending, AvailableAt: time.Now().UTC()})
}

func (m *OutboxMailer) SendWelcome(ctx context.Context, email, name string) error {
	html, err := m.renderer.render("welcome", WelcomeData{Name: name, DashboardURL: m.dashboard})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "Welcome to Outpipe", HTML: html})
}

func (m *OutboxMailer) SendAccountUpdate(ctx context.Context, email, event string) error {
	html, err := m.renderer.render("account-update", AccountUpdateData{Event: event, DashboardURL: m.dashboard})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "Outpipe account update", HTML: html})
}

func (m *OutboxMailer) SendOrganizationInvite(ctx context.Context, email, inviterName, organizationName, role, invitationLink string) error {
	html, err := m.renderer.render("organization-invite", OrganizationInviteData{InviterName: inviterName, OrganizationName: organizationName, Role: role, InvitationLink: invitationLink})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "You’re invited to join " + organizationName + " on Outpipe", HTML: html})
}

func (m *OutboxMailer) SendBillingUpdate(ctx context.Context, organizationID, status string) error {
	to, err := m.resolve(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("resolve billing recipient: %w", err)
	}
	return m.QueueBillingUpdate(ctx, to, status)
}

func (m *OutboxMailer) SendPaymentFailed(ctx context.Context, email, name, planName, amount, billingURL string, attemptsRemaining int, attemptsKnown bool) error {
	html, err := m.renderer.render("payment-failed", PaymentFailedData{Name: name, PlanName: planName, Amount: amount, BillingURL: billingURL, AttemptsRemaining: attemptsRemaining, AttemptsKnown: attemptsKnown})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "Action required: Outpipe payment failed", HTML: html})
}

func (m *OutboxMailer) SendSubscriptionReset(ctx context.Context, email, name, organizationName, previousPlan, dashboardURL string) error {
	html, err := m.renderer.render("subscription-reset", SubscriptionResetData{Name: name, OrganizationName: organizationName, PreviousPlan: previousPlan, DashboardURL: dashboardURL})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "Your " + organizationName + " subscription changed", HTML: html})
}

func (m *OutboxMailer) SendInvoiceReceipt(ctx context.Context, email, name, organizationName, amount, invoiceURL string) error {
	html, err := m.renderer.render("invoice-receipt", InvoiceReceiptData{Name: name, Organization: organizationName, Amount: amount, InvoiceURL: invoiceURL, DashboardURL: m.dashboard})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "Outpipe payment receipt", HTML: html})
}

func (m *OutboxMailer) QueueBillingUpdate(ctx context.Context, email, status string) error {
	html, err := m.renderer.render("billing-update", BillingUpdateData{Status: status, DashboardURL: m.dashboard})
	if err != nil {
		return err
	}
	return m.enqueue(ctx, Message{To: email, Subject: "Outpipe subscription update", HTML: html})
}

func (m *OutboxMailer) SendContact(ctx context.Context, name, email, topic, message string) error {
	body := fmt.Sprintf("<p>New contact message from <strong>%s</strong> (%s).</p><p>Topic: %s</p><p>%s</p>", html.EscapeString(name), html.EscapeString(email), html.EscapeString(topic), html.EscapeString(message))
	return m.enqueue(ctx, Message{To: m.support, Subject: "Outpipe contact: " + topic, HTML: body})
}

func (m *OutboxMailer) SendBugReport(ctx context.Context, name, email, category, summary, reproduction, expected, actual string) error {
	body := fmt.Sprintf("<p>Bug report from <strong>%s</strong> (%s).</p><p>Category: %s</p><h2>%s</h2><h3>Steps to reproduce</h3><p>%s</p><h3>Expected</h3><p>%s</p><h3>Actual</h3><p>%s</p>", html.EscapeString(name), html.EscapeString(email), html.EscapeString(category), html.EscapeString(summary), html.EscapeString(reproduction), html.EscapeString(expected), html.EscapeString(actual))
	return m.enqueue(ctx, Message{To: m.support, Subject: "Outpipe bug report: " + summary, HTML: body})
}

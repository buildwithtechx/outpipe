package mail

import (
	"context"
	"fmt"
)

type BillingMailer struct {
	client    *ZeptoClient
	resolve   func(context.Context, string) (string, error)
	renderer  *templateRenderer
	dashboard string
}

func NewBillingMailer(client *ZeptoClient, resolve func(context.Context, string) (string, error), dashboardURL string) (*BillingMailer, error) {

	if client == nil || resolve == nil {
		return nil, fmt.Errorf("zepto client and recipient resolver are required")
	}

	renderer, err := newTemplateRenderer()

	if err != nil {
		return nil, err
	}

	return &BillingMailer{client: client, resolve: resolve, renderer: renderer, dashboard: dashboardURL}, nil
}

func (m *BillingMailer) SendBillingUpdate(ctx context.Context, organizationID, status string) error {
	to, err := m.resolve(ctx, organizationID)

	if err != nil {
		return fmt.Errorf("resolve billing recipient: %w", err)
	}

	html, err := m.renderer.render("billing-update", BillingUpdateData{Status: status, DashboardURL: m.dashboard})

	if err != nil {
		return err
	}

	return m.client.Send(ctx, Message{To: to, Subject: "Outpipe subscription update", HTML: html})
}

func (m *BillingMailer) SendPaymentFailed(ctx context.Context, email, name, planName, amount, billingURL string, attemptsRemaining int, attemptsKnown bool) error {
	html, err := m.renderer.render("payment-failed", PaymentFailedData{Name: name, PlanName: planName, Amount: amount, BillingURL: billingURL, AttemptsRemaining: attemptsRemaining, AttemptsKnown: attemptsKnown})

	if err != nil {
		return err
	}

	return m.client.Send(ctx, Message{To: email, Subject: "Action required: Outpipe payment failed", HTML: html})
}

func (m *BillingMailer) SendSubscriptionReset(ctx context.Context, email, name, organizationName, previousPlan, dashboardURL string) error {
	html, err := m.renderer.render("subscription-reset", SubscriptionResetData{Name: name, OrganizationName: organizationName, PreviousPlan: previousPlan, DashboardURL: dashboardURL})

	if err != nil {
		return err
	}

	subject := "Your " + organizationName + " subscription changed"
	return m.client.Send(ctx, Message{To: email, Subject: subject, HTML: html})
}

func (m *BillingMailer) SendInvoiceReceipt(ctx context.Context, email, name, organizationName, amount, invoiceURL string) error {
	html, err := m.renderer.render("invoice-receipt", InvoiceReceiptData{Name: name, Organization: organizationName, Amount: amount, InvoiceURL: invoiceURL, DashboardURL: m.dashboard})
	if err != nil {
		return err
	}
	return m.client.Send(ctx, Message{To: email, Subject: "Outpipe payment receipt", HTML: html})
}

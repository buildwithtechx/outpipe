package mail

import (
	"strings"
	"testing"
)

func TestTemplateRendererEscapesHTMLData(t *testing.T) {
	renderer, err := newTemplateRenderer()

	if err != nil {
		t.Fatalf("create template renderer: %v", err)
	}

	html, err := renderer.render("account-update", AccountUpdateData{Event: "<disabled>", DashboardURL: "https://localhost/dashboard"})

	if err != nil {
		t.Fatalf("render account update: %v", err)
	}

	if strings.Contains(html, "<disabled>") || !strings.Contains(html, "&lt;disabled&gt;") {
		t.Fatalf("expected escaped event in html: %s", html)
	}
}

func TestTemplateRendererSupportsTransactionalTemplates(t *testing.T) {
	renderer, err := newTemplateRenderer()

	if err != nil {
		t.Fatalf("create template renderer: %v", err)
	}

	templates := map[string]any{
		"welcome":             WelcomeData{Name: "Ada", DashboardURL: "https://localhost"},
		"account-update":      AccountUpdateData{Event: "deleted"},
		"billing-update":      BillingUpdateData{Status: "active"},
		"organization-invite": OrganizationInviteData{InviterName: "Ada", OrganizationName: "Acme", Role: "member", InvitationLink: "https://localhost/invite"},
		"payment-failed":      PaymentFailedData{Name: "Ada", PlanName: "Beam", Amount: "$10", BillingURL: "https://localhost/billing", AttemptsRemaining: 1, AttemptsKnown: true},
		"subscription-reset":  SubscriptionResetData{Name: "Ada", OrganizationName: "Acme", PreviousPlan: "Beam", DashboardURL: "https://localhost"},
	}
	for name, data := range templates {

		if _, err := renderer.render(name, data); err != nil {
			t.Errorf("render %s: %v", name, err)
		}
	}
}

package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"outpipe.dev/outpipe/internal/models"
)

func TestPolarCheckoutSelectsProductByInterval(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 4096)
		n, _ := r.Body.Read(body)
		requests = append(requests, strings.TrimSpace(string(body[:n])))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://polar.sh/checkout/1"}`))
	}))
	defer server.Close()

	client, err := NewPolar(PolarConfig{
		BaseURL:     server.URL,
		AccessToken: "token",
		ProductIDs: map[string]string{
			"link": "prod_link_monthly",
		},
		YearlyProductIDs: map[string]string{
			"link": "prod_link_yearly",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	monthly := models.Plan{Key: "link", BillingInterval: models.BillingIntervalMonth}

	if _, err := client.Checkout(context.Background(), monthly, "org-1"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(requests[0], `"product_id":"prod_link_monthly"`) {
		t.Errorf("monthly checkout used wrong product: %s", requests[0])
	}

	if !strings.Contains(requests[0], `"billing_interval":"month"`) {
		t.Errorf("monthly checkout missing interval metadata: %s", requests[0])
	}

	yearly := models.Plan{Key: "link", BillingInterval: models.BillingIntervalYear}

	if _, err := client.Checkout(context.Background(), yearly, "org-1"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(requests[1], `"product_id":"prod_link_yearly"`) {
		t.Errorf("yearly checkout used wrong product: %s", requests[1])
	}

	if !strings.Contains(requests[1], `"billing_interval":"year"`) {
		t.Errorf("yearly checkout missing interval metadata: %s", requests[1])
	}
}

func TestPolarCheckoutRejectsMissingYearlyProduct(t *testing.T) {
	client, err := NewPolar(PolarConfig{
		BaseURL:          "https://api.polar.test",
		AccessToken:      "token",
		ProductIDs:       map[string]string{"link": "prod_link_monthly"},
		YearlyProductIDs: map[string]string{},
	})

	if err != nil {
		t.Fatal(err)
	}

	yearly := models.Plan{Key: "link", BillingInterval: models.BillingIntervalYear}

	if _, err := client.Checkout(context.Background(), yearly, "org-1"); err == nil {
		t.Fatal("expected missing yearly product error")
	}
}

func TestAnnualPrice(t *testing.T) {
	tests := []struct {
		monthly int64
		want    int64
	}{
		{700, 7000},
		{1500, 15000},
		{12000, 120000},
		{0, 0},
	}

	for _, tt := range tests {

		if got := annualPrice(tt.monthly); got != tt.want {
			t.Errorf("annualPrice(%d) = %d, want %d", tt.monthly, got, tt.want)
		}
	}
}

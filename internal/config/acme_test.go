package config

import "testing"

func TestACMERequiresOptInAndContact(t *testing.T) {
	cfg := AppConfig{Port: "8080", PublicAPIURL: "http://localhost:8080", DashboardURL: "http://localhost:3000"}
	if err := validateAPIApp(cfg); err != nil {
		t.Fatal(err)
	}
	cfg.ACMEEnabled = true
	if err := validateAPIApp(cfg); err == nil {
		t.Fatal("enabled ACME accepted an empty contact email")
	}
	cfg.ACMEEmail = "owner@example.com"
	if err := validateAPIApp(cfg); err != nil {
		t.Fatal(err)
	}
}

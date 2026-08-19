package services

import (
	"testing"

	"outpipe.dev/outpipe/internal/models"
)

func TestNormalizeBillingInterval(t *testing.T) {
	tests := []struct {
		input string
		want  string
		isErr bool
	}{
		{"", models.BillingIntervalMonth, false},
		{models.BillingIntervalMonth, models.BillingIntervalMonth, false},
		{models.BillingIntervalYear, models.BillingIntervalYear, false},
		{"quarterly", "", true},
		{"weekly", "", true},
	}

	for _, tt := range tests {
		got, err := normalizeBillingInterval(tt.input)

		if tt.isErr {

			if err == nil {
				t.Errorf("normalizeBillingInterval(%q): expected error", tt.input)
			}

			continue
		}

		if err != nil {
			t.Errorf("normalizeBillingInterval(%q): %v", tt.input, err)
			continue
		}

		if got != tt.want {
			t.Errorf("normalizeBillingInterval(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidHostname(t *testing.T) {
	valid := []string{"example.com", "api.example.com", "a1.example-2.com"}

	for _, hostname := range valid {

		if !validHostname(hostname) {
			t.Fatalf("expected hostname %q to be valid", hostname)
		}
	}

	invalid := []string{"localhost", "127.0.0.1", "-example.com", "example..com", "example_com"}

	for _, hostname := range invalid {

		if validHostname(hostname) {
			t.Fatalf("expected hostname %q to be invalid", hostname)
		}
	}
}

func TestMemberRoleRank(t *testing.T) {

	if memberRoleRank(models.MemberRoleOwner) <= memberRoleRank(models.MemberRoleAdmin) {
		t.Fatal("owner must outrank admin")
	}

	if memberRoleRank(models.MemberRoleAdmin) <= memberRoleRank(models.MemberRoleMember) {
		t.Fatal("admin must outrank member")
	}

	if memberRoleRank(models.MemberRoleMember) <= memberRoleRank(models.MemberRoleViewer) {
		t.Fatal("member must outrank viewer")
	}
}

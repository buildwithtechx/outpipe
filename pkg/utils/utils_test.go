package utils

import (
	"errors"
	"testing"
)

func TestBase64(t *testing.T) {
	orig := []byte("hello outpipe utils")
	encoded := EncodeBase64(orig)
	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding base64: %v", err)
	}
	if string(decoded) != string(orig) {
		t.Fatalf("expected %s, got %s", string(orig), string(decoded))
	}
}

func TestErrors(t *testing.T) {
	clientErr := NewClientErrorf("not found: %d", 404)
	if !IsClientError(clientErr) {
		t.Fatalf("expected client error")
	}
	if clientErr.Error() != "not found: 404" {
		t.Fatalf("unexpected error message: %s", clientErr.Error())
	}

	authErr := NewAuthorizationError(errors.New("unauthorized"))
	if !IsAuthorizationError(authErr) {
		t.Fatalf("expected authorization error")
	}
	if NewAuthorizationError(nil) != nil {
		t.Fatalf("expected nil for nil input error")
	}
}

func TestMoney(t *testing.T) {
	cases := []struct {
		amount   int64
		currency string
		expected string
	}{
		{1050, "USD", "$10.50"},
		{1050, "EUR", "€10.50"},
		{1050, "GBP", "£10.50"},
		{1050, "NGN", "₦10.50"},
		{1050, "JPY", "¥1050"},
		{1500, "BHD", "1.500 BHD"},
		{2500, "CAD", "25.00 CAD"},
	}

	for _, tc := range cases {
		got := FormatMinorAmount(tc.amount, tc.currency)
		if got != tc.expected {
			t.Errorf("FormatMinorAmount(%d, %q) = %q; want %q", tc.amount, tc.currency, got, tc.expected)
		}
	}
}

func TestPointer(t *testing.T) {
	val := 42
	p := Ptr(val)
	if p == nil || *p != 42 {
		t.Fatalf("expected pointer to 42, got %v", p)
	}

	if Deref(p, 0) != 42 {
		t.Fatalf("expected 42, got %d", Deref(p, 0))
	}

	var nilPtr *int
	if Deref(nilPtr, 99) != 99 {
		t.Fatalf("expected fallback 99, got %d", Deref(nilPtr, 99))
	}
	if DerefOrZero(nilPtr) != 0 {
		t.Fatalf("expected 0, got %d", DerefOrZero(nilPtr))
	}
	if DerefOrZero(p) != 42 {
		t.Fatalf("expected 42, got %d", DerefOrZero(p))
	}
}

func TestSlices(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	if !Contains(nums, 3) || Contains(nums, 10) {
		t.Fatalf("Contains failed for slice")
	}

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	if len(evens) != 2 || evens[0] != 2 || evens[1] != 4 {
		t.Fatalf("Filter failed: %v", evens)
	}

	doubled := Map(nums, func(n int) int { return n * 2 })
	if len(doubled) != 5 || doubled[0] != 2 || doubled[4] != 10 {
		t.Fatalf("Map failed: %v", doubled)
	}

	raw := []string{"  alpha ", "", "   ", "beta", "gamma  "}
	compacted := CompactStrings(raw)
	if len(compacted) != 3 || compacted[0] != "alpha" || compacted[1] != "beta" || compacted[2] != "gamma" {
		t.Fatalf("CompactStrings failed: %v", compacted)
	}
}

func TestStrings(t *testing.T) {
	if FirstNonEmpty("", "  ", "found", "ignored") != "found" {
		t.Fatalf("FirstNonEmpty failed")
	}
	if FirstNonEmpty("", "   ") != "" {
		t.Fatalf("FirstNonEmpty failed on empty strings")
	}

	if Truncate("hello", 10) != "hello" {
		t.Fatalf("Truncate short string failed")
	}
	if Truncate("hello world", 8) != "hello..." {
		t.Fatalf("Truncate long string failed: got %s", Truncate("hello world", 8))
	}
	if Truncate("hello", 2) != "he" {
		t.Fatalf("Truncate very short maxLen failed: got %s", Truncate("hello", 2))
	}
	if Truncate("hello", 0) != "" {
		t.Fatalf("Truncate zero maxLen failed")
	}
}

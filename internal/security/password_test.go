package security

import "testing"

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")

	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if !VerifyPassword("correct horse battery staple", hash) {
		t.Fatal("expected password to verify")
	}

	if VerifyPassword("wrong password", hash) {
		t.Fatal("expected wrong password to fail")
	}
}

package auth

import (
	"testing"
)

func TestEncryptDecryptSecret(t *testing.T) {
	key := "12345678901234567890123456789012"
	secret := "sk_test_paystack_secret_key_12345"

	encrypted, err := EncryptSecret(secret, key)

	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}

	if encrypted == secret {
		t.Fatal("encrypted string should not match secret")
	}

	decrypted, err := DecryptSecret(encrypted, key)

	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}

	if decrypted != secret {
		t.Fatalf("expected decrypted secret %q, got %q", secret, decrypted)
	}
}

package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaystackInitializeTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		if request.URL.Path != "/transaction/initialize" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("missing authorization header")
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{"status":true,"message":"ok","data":{"authorization_url":"https://paystack.test/checkout","access_code":"access","reference":"reference"}}`))
	}))
	defer server.Close()

	client, err := NewPaystack(PaystackConfig{BaseURL: server.URL, SecretKey: "secret"})

	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	transaction, err := client.InitializeTransaction(context.Background(), "user@example.com", 1000, nil)

	if err != nil {
		t.Fatalf("initialize transaction: %v", err)
	}

	if transaction.Reference != "reference" {
		t.Fatalf("unexpected reference: %s", transaction.Reference)
	}
}

func TestPaystackVerifyWebhook(t *testing.T) {
	payload := []byte(`{"event":"charge.success"}`)
	hasher := hmac.New(sha512.New, []byte("secret"))
	_, _ = hasher.Write(payload)
	signature := hex.EncodeToString(hasher.Sum(nil))
	client, err := NewPaystack(PaystackConfig{SecretKey: "secret"})

	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if !client.VerifyWebhook(payload, signature) {
		t.Fatal("expected webhook signature to verify")
	}
}

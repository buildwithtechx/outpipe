package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type PaystackConfig struct {
	BaseURL    string
	SecretKey  string
	HTTPClient *http.Client
}

func (c *PaystackClient) Portal(context.Context, string) (string, error) {
	return "", fmt.Errorf("paystack does not provide a hosted customer portal")
}

func (c *PaystackClient) Cancel(ctx context.Context, subscriptionID string) error {
	_, err := c.request(ctx, http.MethodPost, "/subscription/disable", map[string]any{"code": subscriptionID, "token": subscriptionID})
	return err
}

func (c *PaystackClient) Resume(ctx context.Context, subscriptionID string) error {
	_, err := c.request(ctx, http.MethodPost, "/subscription/enable", map[string]any{"code": subscriptionID, "token": subscriptionID})
	return err
}

type PaystackClient struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
}

type Transaction struct {
	AuthorizationURL string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
	Status           string `json:"status"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
}

func NewPaystack(cfg PaystackConfig) (*PaystackClient, error) {

	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("paystack secret key is required")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	if baseURL == "" {
		baseURL = "https://api.paystack.co"
	}

	client := cfg.HTTPClient

	if client == nil {
		client = http.DefaultClient
	}

	return &PaystackClient{baseURL: baseURL, secretKey: cfg.SecretKey, httpClient: client}, nil
}

func (c *PaystackClient) InitializeTransaction(ctx context.Context, email string, amount int64, metadata map[string]any) (Transaction, error) {
	return c.request(ctx, http.MethodPost, "/transaction/initialize", map[string]any{
		"email": email, "amount": amount, "channels": []string{"card"}, "metadata": metadata,
	})
}

func (c *PaystackClient) ChargeAuthorization(ctx context.Context, authorizationCode string, email string, amount int64, reference string, metadata map[string]any) (Transaction, error) {
	return c.request(ctx, http.MethodPost, "/transaction/charge_authorization", map[string]any{
		"authorization_code": authorizationCode, "email": email, "amount": amount,
		"reference": reference, "metadata": metadata,
	})
}

func (c *PaystackClient) VerifyTransaction(ctx context.Context, reference string) (Transaction, error) {
	return c.request(ctx, http.MethodGet, "/transaction/verify/"+url.PathEscape(reference), nil)
}

func (c *PaystackClient) VerifyWebhook(payload []byte, signature string) bool {
	hasher := hmac.New(sha512.New, []byte(c.secretKey))
	_, _ = hasher.Write(payload)
	expected, err := hex.DecodeString(signature)

	if err != nil {
		return false
	}

	return hmac.Equal(hasher.Sum(nil), expected)
}

func (c *PaystackClient) request(ctx context.Context, method string, path string, payload any) (Transaction, error) {
	var body io.Reader

	if payload != nil {
		encoded, err := json.Marshal(payload)

		if err != nil {
			return Transaction{}, fmt.Errorf("encode paystack request: %w", err)
		}

		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)

	if err != nil {
		return Transaction{}, fmt.Errorf("create paystack request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.secretKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)

	if err != nil {
		return Transaction{}, fmt.Errorf("send paystack request: %w", err)
	}

	defer response.Body.Close()
	var envelope struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Data    Transaction `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return Transaction{}, fmt.Errorf("decode paystack response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !envelope.Status {
		return Transaction{}, fmt.Errorf("paystack request failed: %s", envelope.Message)
	}

	return envelope.Data, nil
}

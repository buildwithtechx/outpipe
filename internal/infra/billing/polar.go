package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"outpipe.dev/outpipe/internal/models"
)

type PolarConfig struct {
	BaseURL     string
	AccessToken string
	HTTPClient  *http.Client
}

func (c *PolarClient) Checkout(ctx context.Context, plan models.Plan, organizationID string) (string, error) {
	var response struct {
		URL string `json:"url"`
	}
	if err := c.Request(ctx, http.MethodPost, "/v1/checkouts", map[string]any{"product_id": plan.Key, "metadata": map[string]string{"organization_id": organizationID}}, &response); err != nil {
		return "", err
	}

	if response.URL == "" {
		return "", fmt.Errorf("polar checkout returned no url")
	}

	return response.URL, nil
}

func (c *PolarClient) Portal(ctx context.Context, customerID string) (string, error) {
	var response struct {
		URL string `json:"url"`
	}
	if err := c.Request(ctx, http.MethodPost, "/v1/customer-portal/sessions", map[string]string{"customer_id": customerID}, &response); err != nil {
		return "", err
	}

	return response.URL, nil
}

func (c *PolarClient) Cancel(ctx context.Context, subscriptionID string) error {
	return c.Request(ctx, http.MethodPost, "/v1/subscriptions/"+subscriptionID+"/cancel", nil, nil)
}

func (c *PolarClient) Resume(ctx context.Context, subscriptionID string) error {
	return c.Request(ctx, http.MethodPost, "/v1/subscriptions/"+subscriptionID+"/uncancel", nil, nil)
}

type PolarClient struct {
	baseURL     string
	accessToken string
	httpClient  *http.Client
}

func NewPolar(cfg PolarConfig) (*PolarClient, error) {

	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("polar access token is required")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	if baseURL == "" {
		baseURL = "https://sandbox-api.polar.sh"
	}

	client := cfg.HTTPClient

	if client == nil {
		client = http.DefaultClient
	}

	return &PolarClient{baseURL: baseURL, accessToken: cfg.AccessToken, httpClient: client}, nil
}

func (c *PolarClient) Request(ctx context.Context, method string, path string, payload any, responseBody any) error {
	var body io.Reader

	if payload != nil {
		encoded, err := json.Marshal(payload)

		if err != nil {
			return fmt.Errorf("encode polar request: %w", err)
		}

		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)

	if err != nil {
		return fmt.Errorf("create polar request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.accessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)

	if err != nil {
		return fmt.Errorf("send polar request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))

		if readErr != nil {
			return fmt.Errorf("polar request failed with status %d: %w", response.StatusCode, readErr)
		}

		return fmt.Errorf("polar request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	if responseBody == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("decode polar response: %w", err)
	}

	return nil
}

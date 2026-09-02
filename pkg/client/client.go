package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"outpipe.dev/outpipe/internal/infra/httpclient"
)

type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	if baseURL == "" {
		return nil, fmt.Errorf("client base url is required")
	}

	httpClient := cfg.HTTPClient

	if httpClient == nil {
		httpClient = httpclient.New(0)
	}

	return &Client{baseURL: baseURL, apiKey: cfg.APIKey, httpClient: httpClient}, nil
}

func (c *Client) Do(ctx context.Context, method string, path string, requestBody any, responseBody any) error {
	var body io.Reader

	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)

		if err != nil {
			return fmt.Errorf("encode api request: %w", err)
		}

		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/"+strings.TrimLeft(path, "/"), body)

	if err != nil {
		return fmt.Errorf("create api request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	response, err := c.httpClient.Do(request)

	if err != nil {
		return fmt.Errorf("send api request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))

		if readErr != nil {
			return fmt.Errorf("api request failed with status %d: %w", response.StatusCode, readErr)
		}

		return fmt.Errorf("api request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	if responseBody == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("decode api response: %w", err)
	}

	return nil
}

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	APIPrefix  string
}
type Client struct {
	baseURL    string
	apiKey     string
	apiPrefix  string
	httpClient *http.Client
}
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("outpipe API request failed with status %d: %s", e.StatusCode, e.Message)
}

func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	prefix := strings.Trim(cfg.APIPrefix, "/")
	if prefix == "" {
		prefix = "api/v1"
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{baseURL: baseURL, apiKey: cfg.APIKey, apiPrefix: prefix, httpClient: httpClient}, nil
}

func (c *Client) Health(ctx context.Context) (map[string]any, error) {
	var output map[string]any
	return output, c.do(ctx, http.MethodGet, "/healthz", nil, &output)
}
func (c *Client) Ready(ctx context.Context) (map[string]any, error) {
	var output map[string]any
	return output, c.do(ctx, http.MethodGet, "/readyz", nil, &output)
}
func (c *Client) Organizations(ctx context.Context) ([]map[string]any, error) {
	return requestList(ctx, c, "/organizations", nil)
}
func (c *Client) Organization(ctx context.Context, id string) (map[string]any, error) {
	return requestMap(ctx, c, http.MethodGet, "/organizations/"+url.PathEscape(id), nil)
}
func (c *Client) Tunnels(ctx context.Context, organizationID string) ([]map[string]any, error) {
	return requestList(ctx, c, "/organizations/"+url.PathEscape(organizationID)+"/tunnels", nil)
}
func (c *Client) CreateTunnel(ctx context.Context, organizationID string, tunnel map[string]any) (map[string]any, error) {
	return requestMap(ctx, c, http.MethodPost, "/organizations/"+url.PathEscape(organizationID)+"/tunnels", tunnel)
}
func (c *Client) Tunnel(ctx context.Context, tunnelID string) (map[string]any, error) {
	return requestMap(ctx, c, http.MethodGet, "/tunnels/"+url.PathEscape(tunnelID), nil)
}
func (c *Client) SetTunnelStatus(ctx context.Context, tunnelID, status string) error {
	_, err := requestMap(ctx, c, http.MethodPatch, "/tunnels/"+url.PathEscape(tunnelID)+"/status", map[string]string{"status": status})
	return err
}
func (c *Client) RevokeTunnel(ctx context.Context, tunnelID string) error {
	return requestNoContent(ctx, c, http.MethodDelete, "/tunnels/"+url.PathEscape(tunnelID), nil)
}

func (c *Client) do(ctx context.Context, method, path string, input any, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 32*1024*1024))
	if readErr != nil {
		return fmt.Errorf("read response: %w", readErr)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &APIError{StatusCode: response.StatusCode, Message: apiErrorMessage(data), Body: data}
	}
	if output == nil || response.StatusCode == http.StatusNoContent || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func apiErrorMessage(data []byte) string {
	var payload struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err == nil {
		if strings.TrimSpace(payload.Message) != "" {
			return payload.Message
		}
		if strings.TrimSpace(payload.Error) != "" {
			return payload.Error
		}
	}
	return strings.TrimSpace(string(data))
}

func requestMap(ctx context.Context, c *Client, method, path string, input any) (map[string]any, error) {
	var output map[string]any
	return output, c.do(ctx, method, "/"+c.apiPrefix+path, input, &output)
}
func requestList(ctx context.Context, c *Client, path string, input any) ([]map[string]any, error) {
	var output []map[string]any
	return output, c.do(ctx, http.MethodGet, "/"+c.apiPrefix+path, input, &output)
}
func requestNoContent(ctx context.Context, c *Client, method, path string, input any) error {
	return c.do(ctx, method, "/"+c.apiPrefix+path, input, nil)
}

package dns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/validation"
)

const cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

type CloudflareProvider struct {
	token   string
	client  *http.Client
	baseURL string
}

func NewCloudflareProvider(token string, client *http.Client) (*CloudflareProvider, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("cloudflare API token is required")
	}
	if client == nil {
		client = validation.NewSafeHTTPClient(30 * time.Second)
	}
	return &CloudflareProvider{token: token, client: client, baseURL: cloudflareAPIBase}, nil
}

func (p *CloudflareProvider) Name() string { return "cloudflare" }

type cloudflareEnvelope struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (p *CloudflareProvider) call(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode cloudflare request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("create cloudflare request: %w", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Authorization", "Bearer "+p.token)
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("cloudflare request failed: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read cloudflare response: %w", err)
	}
	var envelope cloudflareEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode cloudflare response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || !envelope.Success {
		message := fmt.Sprintf("cloudflare returned HTTP %d", response.StatusCode)
		if len(envelope.Errors) > 0 {
			message = fmt.Sprintf("%s: %s", message, envelope.Errors[0].Message)
		}
		return nil, errors.New(message)
	}
	return raw, nil
}

func validRecordName(name string) bool {
	if name == "" || len(name) > 253 {
		return false
	}
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' && c != '.' && c != '_' {
			return false
		}
	}
	return true
}

func (p *CloudflareProvider) zoneID(ctx context.Context, fqdn string) (string, error) {
	name := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	for {
		raw, err := p.call(ctx, http.MethodGet, "/zones?name="+name, nil)
		if err != nil {
			return "", err
		}
		var result struct {
			cloudflareEnvelope
			Result []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"result"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			return "", fmt.Errorf("decode cloudflare zone: %w", err)
		}
		for _, zone := range result.Result {
			if strings.EqualFold(zone.Name, name) {
				return zone.ID, nil
			}
		}
		parent := strings.Index(name, ".")
		if parent < 0 {
			return "", fmt.Errorf("no cloudflare zone covers %s", fqdn)
		}
		name = name[parent+1:]
	}
}

func (p *CloudflareProvider) Present(ctx context.Context, fqdn, value string) error {
	fqdn = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	if !validRecordName(fqdn) || value == "" || len(value) > 1024 {
		return fmt.Errorf("dns record name and value are required")
	}
	zone, err := p.zoneID(ctx, fqdn)
	if err != nil {
		return err
	}
	raw, err := p.call(ctx, http.MethodGet, "/zones/"+zone+"/dns_records?type=TXT&name="+fqdn, nil)
	if err != nil {
		return err
	}
	var existing struct {
		cloudflareEnvelope
		Result []struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &existing); err != nil {
		return fmt.Errorf("decode cloudflare records: %w", err)
	}
	for _, record := range existing.Result {
		if record.Content == value {
			return nil
		}
		if _, err := p.call(ctx, http.MethodDelete, "/zones/"+zone+"/dns_records/"+record.ID, nil); err != nil {
			return err
		}
	}
	_, err = p.call(ctx, http.MethodPost, "/zones/"+zone+"/dns_records", map[string]any{"type": "TXT", "name": fqdn, "content": value, "ttl": 120})
	return err
}

func (p *CloudflareProvider) Cleanup(ctx context.Context, fqdn string) error {
	fqdn = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(fqdn)), ".")
	if !validRecordName(fqdn) {
		return fmt.Errorf("dns record name is required")
	}
	zone, err := p.zoneID(ctx, fqdn)
	if err != nil {
		return err
	}
	raw, err := p.call(ctx, http.MethodGet, "/zones/"+zone+"/dns_records?type=TXT&name="+fqdn, nil)
	if err != nil {
		return err
	}
	var existing struct {
		cloudflareEnvelope
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &existing); err != nil {
		return fmt.Errorf("decode cloudflare records: %w", err)
	}
	for _, record := range existing.Result {
		if _, err := p.call(ctx, http.MethodDelete, "/zones/"+zone+"/dns_records/"+record.ID, nil); err != nil {
			return err
		}
	}
	return nil
}

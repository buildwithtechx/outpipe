package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/infra/httpclient"
	"outpipe.dev/outpipe/internal/models"
)

type usageRecorder struct {
	url    string
	secret string
	client *http.Client
}

func newUsageRecorder(baseURL, secret string) *usageRecorder {
	return &usageRecorder{url: strings.TrimRight(baseURL, "/") + "/internal/usage", secret: secret, client: httpclient.New(0)}
}

func (r *usageRecorder) Record(ctx context.Context, measurement engine.UsageMeasurement) error {

	if r == nil || r.secret == "" || measurement.OrganizationID == "" {
		return nil
	}

	var tunnelID *string

	if measurement.TunnelID != "" {
		value := measurement.TunnelID
		tunnelID = &value
	}

	event := models.UsageEvent{
		OrganizationID: measurement.OrganizationID,
		TunnelID:       tunnelID,
		EventType:      measurement.EventType,
		Bytes:          measurement.Bytes,
		Connections:    measurement.Connections,
		Method:         measurement.Method,
		Path:           measurement.Path,
		StatusCode:     measurement.StatusCode,
		DurationMillis: measurement.DurationMillis,
		ResponseBytes:  measurement.ResponseBytes,
		ClientIP:       measurement.ClientIP,
	}
	body, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("encode usage event: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(body))

	if err != nil {
		return fmt.Errorf("create usage request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", r.secret)
	response, err := r.client.Do(request)

	if err != nil {
		return fmt.Errorf("send usage event: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("usage request failed with status %d", response.StatusCode)
	}

	return nil
}

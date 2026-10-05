package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/validation"
)

const (
	maxReplayBodySize     = 10 * 1024 * 1024
	maxReplayResponseBody = 5 * 1024 * 1024
	replayTimeout         = 15 * time.Second
)

func (s *ObservabilityService) ReplayRequest(ctx context.Context, input models.ReplayRequestInput) (*models.ReplayResponseOutput, error) {
	if err := validation.ValidateWebhookURL(input.URL); err != nil {
		return nil, fmt.Errorf("validate replay target: %w", err)
	}
	if input.URL == "" {
		return nil, fmt.Errorf("target URL is required")
	}
	if input.Method == "" {
		input.Method = http.MethodGet
	}

	parsedURL, err := url.Parse(input.URL)
	if err != nil {
		return nil, fmt.Errorf("parse target URL: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme %q; only http and https are permitted", scheme)
	}

	if len(input.RequestBody) > maxReplayBodySize {
		return nil, fmt.Errorf("request body exceeds maximum allowed size of 10MB")
	}

	var bodyReader io.Reader
	if input.Method != http.MethodGet && input.Method != http.MethodHead && len(input.RequestBody) > 0 {
		bodyReader = strings.NewReader(input.RequestBody)
	}

	reqCtx, cancel := context.WithTimeout(ctx, replayTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, strings.ToUpper(input.Method), input.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("construct replay request: %w", err)
	}

	for k, v := range input.Headers {
		lowerKey := strings.ToLower(k)
		if lowerKey == "host" || lowerKey == "content-length" {
			continue
		}
		httpReq.Header.Set(k, v)
	}

	client := validation.NewSafeHTTPClient(replayTimeout)
	defer client.CloseIdleConnections()

	start := time.Now()
	resp, err := client.Do(httpReq)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("execute replay request: %w", err)
	}
	defer resp.Body.Close()

	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxReplayResponseBody))
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &models.ReplayResponseOutput{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Headers:    respHeaders,
		Body:       string(bodyBytes),
		DurationMs: duration,
	}, nil
}

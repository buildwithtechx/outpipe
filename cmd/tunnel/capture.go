package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/relay"
)

type captureRecorder struct {
	resolver    *relay.InternalTunnelResolver
	baseURL     string
	secret      string
	client      *http.Client
	mu          sync.Mutex
	policies    map[string]capturePolicy
	policyQueue chan capturePolicyRequest
	queue       chan engine.RequestCapture
	wait        sync.WaitGroup
}

func (r *captureRecorder) enabledPolicy(ctx context.Context, tunnelID, orgID string) (bool, error) {
	policy, err := r.resolver.Resolve(ctx, tunnelID)
	if err != nil {
		return false, fmt.Errorf("resolve capture policy: %w", err)
	}
	return policy.OrganizationID == orgID && policy.Status != "revoked" && policy.CaptureEnabled, nil
}

func (r *captureRecorder) sendCapture(ctx context.Context, capture engine.RequestCapture) error {
	body, err := json.Marshal(capture)
	if err != nil {
		return fmt.Errorf("encode capture: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(r.baseURL, "/")+"/internal/captures", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create capture request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", r.secret)
	client := *r.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send capture: %w", err)
	}
	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("close capture response: %w", err)
	}
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("capture returned status %d", response.StatusCode)
	}
	return nil
}

package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"outpipe.dev/outpipe/internal/infra/httpclient"
)

type InternalAgentAuthenticator struct {
	baseURL string
	secret  string
	client  *http.Client
}

type InternalTunnelResolver struct {
	baseURL string
	secret  string
	client  *http.Client
}

func NewInternalAgentAuthenticator(baseURL, secret string, client *http.Client) (*InternalAgentAuthenticator, error) {

	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("internal api url and secret are required")
	}

	if client == nil {
		client = httpclient.New(0)
	}

	return &InternalAgentAuthenticator{baseURL: strings.TrimRight(baseURL, "/"), secret: secret, client: client}, nil
}

func (a *InternalAgentAuthenticator) Authenticate(ctx context.Context, token string) (AgentIdentity, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/internal/agents/authenticate", nil)

	if err != nil {
		return AgentIdentity{}, fmt.Errorf("create agent authentication request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Internal-Secret", a.secret)
	response, err := a.client.Do(request)

	if err != nil {
		return AgentIdentity{}, fmt.Errorf("authenticate agent with api: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return AgentIdentity{}, fmt.Errorf("agent authentication returned status %d", response.StatusCode)
	}

	var body struct {
		AgentID        string `json:"agentId"`
		OrganizationID string `json:"organizationId"`
		Limits         struct {
			MaxTunnels     int   `json:"maxTunnels"`
			MaxConnections int   `json:"maxConnections"`
			BandwidthBytes int64 `json:"bandwidthBytes"`
		} `json:"limits"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return AgentIdentity{}, fmt.Errorf("decode agent authentication response: %w", err)
	}

	if body.AgentID == "" || body.OrganizationID == "" {
		return AgentIdentity{}, fmt.Errorf("agent authentication response is incomplete")
	}

	return AgentIdentity{AgentID: body.AgentID, OrganizationID: body.OrganizationID, MaxTunnels: body.Limits.MaxTunnels, MaxConnections: body.Limits.MaxConnections, BandwidthBytes: body.Limits.BandwidthBytes}, nil
}

func NewInternalTunnelResolver(baseURL, secret string, client *http.Client) (*InternalTunnelResolver, error) {

	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("internal api url and secret are required")
	}

	if client == nil {
		client = httpclient.New(0)
	}

	return &InternalTunnelResolver{baseURL: strings.TrimRight(baseURL, "/"), secret: secret, client: client}, nil
}

func (r *InternalTunnelResolver) Resolve(ctx context.Context, tunnelID string) (ManagedTunnelPolicy, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, r.baseURL+"/internal/tunnels/"+url.PathEscape(tunnelID)+"/policy", nil)

	if err != nil {
		return ManagedTunnelPolicy{}, fmt.Errorf("create tunnel policy request: %w", err)
	}

	request.Header.Set("X-Internal-Secret", r.secret)
	response, err := r.client.Do(request)

	if err != nil {
		return ManagedTunnelPolicy{}, fmt.Errorf("resolve tunnel policy with api: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ManagedTunnelPolicy{}, fmt.Errorf("tunnel policy returned status %d", response.StatusCode)
	}

	var body struct {
		OrganizationID    string `json:"organizationId"`
		PublicHostname    string `json:"publicHostname"`
		PasswordProtected bool   `json:"passwordProtected"`
		Status            string `json:"status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return ManagedTunnelPolicy{}, fmt.Errorf("decode tunnel policy response: %w", err)
	}

	if body.OrganizationID == "" {
		return ManagedTunnelPolicy{}, fmt.Errorf("tunnel policy response is incomplete")
	}

	return ManagedTunnelPolicy{OrganizationID: body.OrganizationID, PublicHostname: body.PublicHostname, PasswordProtected: body.PasswordProtected, Status: body.Status}, nil
}

func (r *InternalTunnelResolver) VerifyPassword(ctx context.Context, tunnelID, password string) (bool, error) {
	payloadBytes, err := json.Marshal(struct {
		Password string `json:"password"`
	}{Password: password})

	if err != nil {
		return false, fmt.Errorf("encode tunnel password request: %w", err)
	}

	payload := strings.NewReader(string(payloadBytes))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/internal/tunnels/"+url.PathEscape(tunnelID)+"/password", payload)

	if err != nil {
		return false, fmt.Errorf("create tunnel password request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Secret", r.secret)
	response, err := r.client.Do(request)

	if err != nil {
		return false, fmt.Errorf("verify tunnel password with api: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("tunnel password verification returned status %d", response.StatusCode)
	}

	var body struct {
		Valid bool `json:"valid"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("decode tunnel password response: %w", err)
	}

	return body.Valid, nil
}

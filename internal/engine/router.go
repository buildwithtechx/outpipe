package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/pkg/protocol"
)

type RequestRouter struct {
	sessions *SessionRegistry
	timeout  time.Duration
	mu       sync.Mutex
	pending  map[string]pendingRequest
}

type pendingRequest struct {
	tunnelID string
	response chan protocol.HTTPResponse
}

func NewRequestRouter(sessions *SessionRegistry, timeout time.Duration) (*RequestRouter, error) {

	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}

	return &RequestRouter{sessions: sessions, timeout: timeout, pending: make(map[string]pendingRequest)}, nil
}

func (r *RequestRouter) ForwardHTTP(ctx context.Context, tunnelID string, request protocol.HTTPRequest) (protocol.HTTPResponse, error) {
	resolved, ok := r.sessions.Resolve(tunnelID)

	if !ok {
		return protocol.HTTPResponse{}, fmt.Errorf("tunnel %q is not connected", tunnelID)
	}

	tunnelID = resolved
	session, ok := r.sessions.Get(tunnelID)

	if !ok {
		return protocol.HTTPResponse{}, fmt.Errorf("tunnel %q is not connected", tunnelID)
	}

	requestID := uuid.NewString()
	payload, err := json.Marshal(request)

	if err != nil {
		return protocol.HTTPResponse{}, fmt.Errorf("encode http request: %w", err)
	}

	response := make(chan protocol.HTTPResponse, 1)
	r.mu.Lock()
	r.pending[requestID] = pendingRequest{tunnelID: tunnelID, response: response}
	r.mu.Unlock()
	message := protocol.Envelope{Version: protocol.Version, Type: protocol.MessageTypeHTTPRequest, RequestID: requestID, Payload: payload}

	if err := session.Send(ctx, message); err != nil {
		r.remove(requestID)
		return protocol.HTTPResponse{}, fmt.Errorf("send http request: %w", err)
	}

	timer := time.NewTimer(r.timeout)
	defer timer.Stop()

	select {
	case result := <-response:
		return result, nil
	case <-ctx.Done():
		r.remove(requestID)
		return protocol.HTTPResponse{}, fmt.Errorf("http request canceled: %w", ctx.Err())
	case <-timer.C:
		r.remove(requestID)
		return protocol.HTTPResponse{}, fmt.Errorf("http request timed out")
	}
}

func (r *RequestRouter) Handle(message protocol.Envelope) bool {

	if message.Type != protocol.MessageTypeHTTPResponse {
		return false
	}

	var response protocol.HTTPResponse

	if err := protocol.DecodePayload(message, &response); err != nil {
		return false
	}

	r.mu.Lock()
	pending, ok := r.pending[message.RequestID]

	if ok {
		delete(r.pending, message.RequestID)
	}

	r.mu.Unlock()

	if !ok {
		return false
	}

	pending.response <- response
	return true
}

func (r *RequestRouter) RemoveTunnel(tunnelID string) {
	r.mu.Lock()

	for requestID, pending := range r.pending {

		if pending.tunnelID == tunnelID {
			delete(r.pending, requestID)
		}
	}

	r.mu.Unlock()
}

func (r *RequestRouter) OrganizationID(tunnelID string) (string, bool) {
	resolved, ok := r.sessions.Resolve(tunnelID)

	if !ok {
		return "", false
	}

	tunnelID = resolved
	session, ok := r.sessions.Get(tunnelID)

	if !ok {
		return "", false
	}

	return session.OrganizationID, true
}

func (r *RequestRouter) PasswordHash(tunnelID string) (string, bool) {
	resolved, ok := r.sessions.Resolve(tunnelID)

	if !ok {
		return "", false
	}

	session, ok := r.sessions.Get(resolved)

	if !ok {
		return "", false
	}

	return session.PasswordHash, true
}

func (r *RequestRouter) remove(requestID string) {
	r.mu.Lock()
	delete(r.pending, requestID)
	r.mu.Unlock()
}

package http

import (
	"context"
	"fmt"
	"sync"

	"outpipe.dev/outpipe/internal/auth"
)

type fakeOAuthProvider struct{}

func (fakeOAuthProvider) Name() string { return "google" }

func (fakeOAuthProvider) AuthorizeURL(state, _, _ string) string {
	return "https://accounts.example.com/oauth?state=" + state
}

func (fakeOAuthProvider) Exchange(context.Context, string, string, string) (auth.OAuthProfile, error) {
	return auth.OAuthProfile{Provider: "google", Subject: "e2e-google-1", Email: "e2e@example.com", Name: "E2E User", EmailVerified: true}, nil
}

type memoryOAuthStateStore struct {
	mu    sync.Mutex
	state map[string]auth.OAuthState
}

func newMemoryOAuthStateStore() *memoryOAuthStateStore {
	return &memoryOAuthStateStore{state: make(map[string]auth.OAuthState)}
}

func (s *memoryOAuthStateStore) Save(_ context.Context, state string, value auth.OAuthState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[state] = value
	return nil
}

func (s *memoryOAuthStateStore) Take(_ context.Context, state string) (auth.OAuthState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.state[state]

	if !ok {
		return auth.OAuthState{}, fmt.Errorf("oauth state not found")
	}

	delete(s.state, state)
	return value, nil
}

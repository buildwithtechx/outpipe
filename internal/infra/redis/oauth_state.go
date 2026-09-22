package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/auth"
)

type OAuthStateStore struct {
	client *Client
	ttl    time.Duration
}

func NewOAuthStateStore(client *Client, ttl time.Duration) (*OAuthStateStore, error) {

	if client == nil || ttl <= 0 {
		return nil, fmt.Errorf("redis client and positive oauth state ttl are required")
	}

	return &OAuthStateStore{client: client, ttl: ttl}, nil
}

func (s *OAuthStateStore) Save(ctx context.Context, state string, value auth.OAuthState) error {
	body, err := json.Marshal(value)

	if err != nil {
		return fmt.Errorf("encode oauth state: %w", err)
	}

	return s.client.Set(ctx, "outpipe:oauth-state:"+state, body, s.ttl)
}

func (s *OAuthStateStore) Take(ctx context.Context, state string) (auth.OAuthState, error) {
	key := "outpipe:oauth-state:" + state
	body, err := s.client.Get(ctx, key)

	if err != nil {
		return auth.OAuthState{}, fmt.Errorf("get oauth state: %w", err)
	}

	if err := s.client.Delete(ctx, key); err != nil {
		return auth.OAuthState{}, fmt.Errorf("delete oauth state: %w", err)
	}

	var value auth.OAuthState

	if err := json.Unmarshal([]byte(body), &value); err != nil {
		return auth.OAuthState{}, fmt.Errorf("decode oauth state: %w", err)
	}

	return value, nil
}

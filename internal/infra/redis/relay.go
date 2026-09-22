package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RelayAffinity struct{ client *Client }

func NewRelayAffinity(client *Client) (*RelayAffinity, error) {

	if client == nil || client.Raw() == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	return &RelayAffinity{client: client}, nil
}

func (a *RelayAffinity) Claim(ctx context.Context, tunnelID, relayID string, ttl time.Duration) (bool, string, error) {

	if tunnelID == "" || relayID == "" || ttl <= 0 {
		return false, "", fmt.Errorf("tunnel, relay, and positive ttl are required")
	}

	key := "outpipe:relay-owner:" + tunnelID
	owner, err := a.client.Raw().Get(ctx, key).Result()

	if err == nil && owner != relayID {
		return false, owner, nil
	}

	if err != nil && err != redis.Nil {
		return false, "", fmt.Errorf("read relay owner: %w", err)
	}

	if err == nil {
		return true, relayID, a.client.Raw().Expire(ctx, key, ttl).Err()
	}

	ok, err := a.client.Raw().SetNX(ctx, key, relayID, ttl).Result()

	if err != nil {
		return false, "", fmt.Errorf("claim relay owner: %w", err)
	}

	if !ok {
		owner, getErr := a.client.Raw().Get(ctx, key).Result()

		if getErr != nil {
			owner = ""
		}

		return false, owner, nil
	}

	return true, relayID, nil
}

func (a *RelayAffinity) Release(ctx context.Context, tunnelID, relayID string) error {
	key := "outpipe:relay-owner:" + tunnelID
	owner, err := a.client.Raw().Get(ctx, key).Result()

	if err == redis.Nil || owner != relayID {
		return nil
	}

	if err != nil {
		return fmt.Errorf("read relay owner: %w", err)
	}

	return a.client.Raw().Del(ctx, key).Err()
}

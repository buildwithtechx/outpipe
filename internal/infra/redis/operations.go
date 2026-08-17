package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type Operations struct {
	client *Client
}

func NewOperations(client *Client) (*Operations, error) {
	if client == nil || client.Raw() == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	return &Operations{client: client}, nil
}

func (o *Operations) SetPresence(ctx context.Context, organizationID, member string, ttl time.Duration) error {
	return o.client.Set(ctx, presenceKey(organizationID, member), "1", ttl)
}

func (o *Operations) Heartbeat(ctx context.Context, organizationID, member string, ttl time.Duration) error {
	return o.client.Set(ctx, heartbeatKey(organizationID, member), strconv.FormatInt(time.Now().Unix(), 10), ttl)
}

func (o *Operations) AddActiveTunnel(ctx context.Context, organizationID, tunnelID string, ttl time.Duration) error {
	if organizationID == "" || tunnelID == "" || ttl <= 0 {
		return fmt.Errorf("organization, tunnel, and positive ttl are required")
	}
	if err := o.client.Raw().SAdd(ctx, activeIndexKey(organizationID), tunnelID).Err(); err != nil {
		return fmt.Errorf("add active tunnel: %w", err)
	}
	return o.client.Set(ctx, activeTunnelKey(tunnelID), organizationID, ttl)
}

func (o *Operations) RemoveActiveTunnel(ctx context.Context, organizationID, tunnelID string) error {
	if err := o.client.Raw().SRem(ctx, activeIndexKey(organizationID), tunnelID).Err(); err != nil {
		return fmt.Errorf("remove active tunnel index: %w", err)
	}
	return o.client.Delete(ctx, activeTunnelKey(tunnelID))
}

func (o *Operations) ActiveTunnels(ctx context.Context, organizationID string) ([]string, error) {
	values, err := o.client.Raw().SMembers(ctx, activeIndexKey(organizationID)).Result()
	if err != nil {
		return nil, fmt.Errorf("list active tunnels: %w", err)
	}
	return values, nil
}

func (o *Operations) ReconcileActiveTunnels(ctx context.Context, organizationID string) error {
	tunnels, err := o.ActiveTunnels(ctx, organizationID)
	if err != nil {
		return err
	}
	for _, tunnelID := range tunnels {
		exists, err := o.client.Raw().Exists(ctx, activeTunnelKey(tunnelID)).Result()
		if err != nil {
			return fmt.Errorf("check active tunnel: %w", err)
		}
		if exists == 0 {
			if err := o.RemoveActiveTunnel(ctx, organizationID, tunnelID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *Operations) AllowRate(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	if key == "" || limit <= 0 || window <= 0 {
		return false, fmt.Errorf("rate limit key, positive limit, and window are required")
	}
	count, err := o.client.Raw().Incr(ctx, "outpipe:rate:"+key).Result()
	if err != nil {
		return false, fmt.Errorf("increment rate limit: %w", err)
	}
	if count == 1 {
		if err := o.client.Raw().Expire(ctx, "outpipe:rate:"+key, window).Err(); err != nil {
			return false, fmt.Errorf("expire rate limit: %w", err)
		}
	}
	return count <= limit, nil
}

func presenceKey(organizationID, member string) string {
	return "outpipe:presence:" + organizationID + ":" + member
}

func heartbeatKey(organizationID, member string) string {
	return "outpipe:heartbeat:" + organizationID + ":" + member
}

func activeIndexKey(organizationID string) string {
	return "outpipe:active-index:" + organizationID
}

func activeTunnelKey(tunnelID string) string {
	return "outpipe:active:" + tunnelID
}

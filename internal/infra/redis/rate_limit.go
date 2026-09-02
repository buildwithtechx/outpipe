package redis

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if c == nil || c.raw == nil {
		return false, fmt.Errorf("redis client is required")
	}
	if key == "" || limit <= 0 || window <= 0 {
		return false, fmt.Errorf("rate limit parameters are invalid")
	}

	result, err := c.raw.Eval(ctx, `
	local count = redis.call("INCR", KEYS[1])
	if count == 1 then redis.call("PEXPIRE", KEYS[1], ARGV[1]) end
	return count
`, []string{key}, window.Milliseconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("check rate limit: %w", err)
	}

	return result <= int64(limit), nil
}

package locks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	goRedis "github.com/redis/go-redis/v9"
)

type Lease struct {
	client *goRedis.Client
	key    string
	token  string
}

func Acquire(ctx context.Context, client *goRedis.Client, key string, ttl time.Duration) (*Lease, error) {

	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	if key == "" {
		return nil, fmt.Errorf("lock key is required")
	}

	if ttl <= 0 {
		return nil, fmt.Errorf("lock ttl must be positive")
	}

	lease := &Lease{client: client, key: key, token: uuid.NewString()}
	acquired, err := client.SetNX(ctx, key, lease.token, ttl).Result()

	if err != nil {
		return nil, fmt.Errorf("acquire lock: %w", err)
	}

	if !acquired {
		return nil, fmt.Errorf("lock %q is already held", key)
	}

	return lease, nil
}

func (l *Lease) Release(ctx context.Context) error {

	if l == nil || l.client == nil {
		return fmt.Errorf("lock lease is required")
	}

	const script = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`

	if err := l.client.Eval(ctx, script, []string{l.key}, l.token).Err(); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}

	return nil
}

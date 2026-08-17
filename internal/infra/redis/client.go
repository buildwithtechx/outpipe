package redis

import (
	"context"
	"fmt"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type Client struct {
	raw *goRedis.Client
}

func Open(ctx context.Context, cfg Config) (*Client, error) {

	if cfg.Host == "" {
		return nil, fmt.Errorf("redis host is required")
	}

	if cfg.Port == "" {
		return nil, fmt.Errorf("redis port is required")
	}

	raw := goRedis.NewClient(&goRedis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := raw.Ping(ctx).Err(); err != nil {
		raw.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{raw: raw}, nil
}

func (c *Client) Raw() *goRedis.Client {
	return c.raw
}

func (c *Client) Close() error {

	if c == nil || c.raw == nil {
		return nil
	}

	return c.raw.Close()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {

	if c == nil || c.raw == nil {
		return "", fmt.Errorf("redis client is required")
	}

	return c.raw.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {

	if c == nil || c.raw == nil {
		return fmt.Errorf("redis client is required")
	}

	return c.raw.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {

	if c == nil || c.raw == nil {
		return fmt.Errorf("redis client is required")
	}

	return c.raw.Del(ctx, keys...).Err()
}

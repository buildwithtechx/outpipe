package http

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	infraredis "outpipe.dev/outpipe/internal/infra/redis"
)

func requestRateLimit(max int, window time.Duration) fiber.Handler {
	return requestRateLimitBy(max, window, func(c *fiber.Ctx) string { return c.IP() })
}

func requestRateLimitBy(max int, window time.Duration, key func(*fiber.Ctx) string) fiber.Handler {
	type bucket struct {
		started time.Time
		count   int
	}
	var mu sync.Mutex
	buckets := make(map[string]bucket)
	return func(c *fiber.Ctx) error {
		bucketKey := key(c)
		now := time.Now()
		mu.Lock()
		value := buckets[bucketKey]
		if value.started.IsZero() || now.Sub(value.started) >= window {
			value = bucket{started: now}
		}
		value.count++
		buckets[bucketKey] = value
		allowed := value.count <= max
		mu.Unlock()
		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}

func requestRateLimitDistributed(client *infraredis.Client, max int, window time.Duration, key func(*fiber.Ctx) string) fiber.Handler {
	local := requestRateLimitBy(max, window, key)
	return func(c *fiber.Ctx) error {
		if client == nil {
			return local(c)
		}
		allowed, err := client.Allow(c.UserContext(), "outpipe:ratelimit:"+key(c), max, window)
		if err != nil {
			return local(c)
		}
		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}

func authenticatedRateLimitKey(c *fiber.Ctx) string {
	if userID, ok := authenticatedUserID(c); ok {
		return "user:" + userID
	}
	return "ip:" + c.IP()
}

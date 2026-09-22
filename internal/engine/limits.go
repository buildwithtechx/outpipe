package engine

import (
	"fmt"
	"sync"
)

type BandwidthLimiter struct {
	mu    sync.Mutex
	usage map[string]int64
}

func NewBandwidthLimiter() *BandwidthLimiter {
	return &BandwidthLimiter{usage: make(map[string]int64)}
}

func (l *BandwidthLimiter) Consume(key string, limit int64, bytes int64) error {

	if key == "" {
		return fmt.Errorf("bandwidth key is required")
	}

	if bytes < 0 {
		return fmt.Errorf("bandwidth bytes cannot be negative")
	}

	if limit <= 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	current := l.usage[key]

	if current > limit-bytes {
		return fmt.Errorf("bandwidth limit exceeded")
	}

	l.usage[key] = current + bytes
	return nil
}

func (l *BandwidthLimiter) Usage(key string) int64 {
	l.mu.Lock()
	usage := l.usage[key]
	l.mu.Unlock()
	return usage
}

func (l *BandwidthLimiter) Reset(key string) {
	l.mu.Lock()
	delete(l.usage, key)
	l.mu.Unlock()
}

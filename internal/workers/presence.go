package workers

import (
	"context"
	"fmt"
	"time"
)

type PresenceStore interface {
	Set(context.Context, string, any, time.Duration) error
	Delete(context.Context, ...string) error
}

type PresenceJob struct {
	store PresenceStore
	key   string
	value string
	ttl   time.Duration
}

func NewPresenceJob(store PresenceStore, key, value string, ttl time.Duration) (*PresenceJob, error) {

	if store == nil || key == "" || ttl <= 0 {
		return nil, fmt.Errorf("presence store, key, and positive ttl are required")
	}

	return &PresenceJob{store: store, key: key, value: value, ttl: ttl}, nil
}

func (j *PresenceJob) Name() string {
	return "presence"
}

func (j *PresenceJob) Run(ctx context.Context) error {

	if err := j.store.Set(ctx, j.key, j.value, j.ttl); err != nil {
		return fmt.Errorf("refresh presence: %w", err)
	}

	return nil
}

func (j *PresenceJob) Remove(ctx context.Context) error {

	if err := j.store.Delete(ctx, j.key); err != nil {
		return fmt.Errorf("remove presence: %w", err)
	}

	return nil
}

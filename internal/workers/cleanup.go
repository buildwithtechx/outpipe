package workers

import (
	"context"
	"fmt"
	"time"
)

type ExpiryStore interface {
	DeleteExpired(context.Context, time.Time) (int64, error)
}

type CleanupJob struct {
	name   string
	stores []ExpiryStore
	now    func() time.Time
}

func NewCleanupJob(stores ...ExpiryStore) (*CleanupJob, error) {

	if len(stores) == 0 {
		return nil, fmt.Errorf("cleanup stores are required")
	}

	return &CleanupJob{name: "cleanup", stores: append([]ExpiryStore(nil), stores...), now: time.Now}, nil
}

func (j *CleanupJob) Name() string {
	return j.name
}

func (j *CleanupJob) Run(ctx context.Context) error {
	now := j.now()

	for _, store := range j.stores {

		if _, err := store.DeleteExpired(ctx, now); err != nil {
			return fmt.Errorf("delete expired records: %w", err)
		}
	}

	return nil
}

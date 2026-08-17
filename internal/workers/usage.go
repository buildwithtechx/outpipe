package workers

import (
	"context"
	"fmt"
	"time"
)

type UsageAggregator interface {
	Aggregate(context.Context, time.Time) error
}

type UsageJob struct {
	aggregator UsageAggregator
	now        func() time.Time
}

func NewUsageJob(aggregator UsageAggregator) (*UsageJob, error) {

	if aggregator == nil {
		return nil, fmt.Errorf("usage aggregator is required")
	}

	return &UsageJob{aggregator: aggregator, now: time.Now}, nil
}

func (j *UsageJob) Name() string {
	return "usage-aggregation"
}

func (j *UsageJob) Run(ctx context.Context) error {

	if err := j.aggregator.Aggregate(ctx, j.now()); err != nil {
		return fmt.Errorf("aggregate usage: %w", err)
	}

	return nil
}

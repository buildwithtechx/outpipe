package workers

import (
	"context"
	"fmt"
	"time"
)

type RetentionEnforcer interface {
	Enforce(context.Context, time.Time) error
}

type RetentionJob struct {
	enforcer RetentionEnforcer
	now      func() time.Time
}

func NewRetentionJob(enforcer RetentionEnforcer) (*RetentionJob, error) {

	if enforcer == nil {
		return nil, fmt.Errorf("retention enforcer is required")
	}

	return &RetentionJob{enforcer: enforcer, now: time.Now}, nil
}

func (j *RetentionJob) Name() string {
	return "retention-enforcement"
}

func (j *RetentionJob) Run(ctx context.Context) error {

	if err := j.enforcer.Enforce(ctx, j.now()); err != nil {
		return fmt.Errorf("enforce retention: %w", err)
	}

	return nil
}

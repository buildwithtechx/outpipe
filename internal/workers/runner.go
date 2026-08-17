package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Job interface {
	Name() string
	Run(context.Context) error
}

type Runner struct {
	jobs     []Job
	interval time.Duration
	logger   *slog.Logger
	tracker  *StatusTracker
}

func NewRunner(jobs []Job, interval time.Duration, logger *slog.Logger) (*Runner, error) {

	if interval <= 0 {
		return nil, fmt.Errorf("worker interval must be positive")
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &Runner{jobs: append([]Job(nil), jobs...), interval: interval, logger: logger, tracker: NewStatusTracker()}, nil
}

func (r *Runner) Statuses() []JobStatus {

	if r == nil || r.tracker == nil {
		return nil
	}

	return r.tracker.Statuses()
}

func (r *Runner) Tracker() *StatusTracker {

	if r == nil {
		return nil
	}

	return r.tracker
}

func (r *Runner) RunOnce(ctx context.Context) error {

	for _, job := range r.jobs {

		if err := job.Run(ctx); err != nil {
			r.logger.ErrorContext(ctx, "worker failed", "worker", job.Name(), "error", err)
			return fmt.Errorf("run %s: %w", job.Name(), err)
		}
	}

	return nil
}

func (r *Runner) Run(ctx context.Context) error {

	if err := r.RunOnce(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				r.logger.ErrorContext(ctx, "worker cycle failed", "error", err)
			}
		}
	}
}

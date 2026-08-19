package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type DeadLetterHandler interface {
	HandleDeadLetter(ctx context.Context, jobName string, err error, attempts int) error
}

type SlogDeadLetterHandler struct {
	logger *slog.Logger
}

func NewSlogDeadLetterHandler(logger *slog.Logger) *SlogDeadLetterHandler {

	if logger == nil {
		logger = slog.Default()
	}

	return &SlogDeadLetterHandler{logger: logger}
}

func (h *SlogDeadLetterHandler) HandleDeadLetter(ctx context.Context, jobName string, err error, attempts int) error {
	h.logger.ErrorContext(ctx, "job dead-lettered after retries", "job", jobName, "attempts", attempts, "error", err)
	return nil
}

type DeadLetterAlertSender interface {
	AlertJobDeadLetter(ctx context.Context, jobName string, err error, attempts int) error
}

type AlertingDeadLetterHandler struct {
	logger *slog.Logger
	alerts DeadLetterAlertSender
}

func NewAlertingDeadLetterHandler(logger *slog.Logger, alerts DeadLetterAlertSender) (*AlertingDeadLetterHandler, error) {

	if alerts == nil {
		return nil, fmt.Errorf("dead letter alert sender is required")
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &AlertingDeadLetterHandler{logger: logger, alerts: alerts}, nil
}

func (h *AlertingDeadLetterHandler) HandleDeadLetter(ctx context.Context, jobName string, err error, attempts int) error {
	h.logger.ErrorContext(ctx, "job dead-lettered after retries", "job", jobName, "attempts", attempts, "error", err)
	return h.alerts.AlertJobDeadLetter(ctx, jobName, err, attempts)
}

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		BackoffFactor:  2.0,
	}
}

type RetryableJob struct {
	job               Job
	config            RetryConfig
	deadLetterHandler DeadLetterHandler
	tracker           *StatusTracker
	sleep             func(context.Context, time.Duration) error
}

func NewRetryableJob(job Job, cfg RetryConfig, dlh DeadLetterHandler, tracker *StatusTracker) (*RetryableJob, error) {

	if job == nil {
		return nil, fmt.Errorf("job is required")
	}

	if cfg.MaxRetries < 0 {
		return nil, fmt.Errorf("max retries must be non-negative")
	}

	if dlh == nil {
		dlh = NewSlogDeadLetterHandler(nil)
	}

	return &RetryableJob{
		job:               job,
		config:            cfg,
		deadLetterHandler: dlh,
		tracker:           tracker,
		sleep: func(ctx context.Context, d time.Duration) error {
			timer := time.NewTimer(d)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}

		},
	}, nil
}

func (j *RetryableJob) Name() string {
	return j.job.Name()
}

func (j *RetryableJob) Run(ctx context.Context) error {
	now := time.Now()

	if j.tracker != nil {
		j.tracker.RecordStart(j.Name(), now)
	}

	var lastErr error
	backoff := j.config.InitialBackoff

	for attempt := 0; attempt <= j.config.MaxRetries; attempt++ {

		if err := ctx.Err(); err != nil {
			return err
		}

		err := j.job.Run(ctx)

		if err == nil {

			if j.tracker != nil {
				j.tracker.RecordSuccess(j.Name(), time.Now())
			}

			return nil
		}

		lastErr = err

		if attempt < j.config.MaxRetries {

			if j.tracker != nil {
				j.tracker.RecordFailure(j.Name(), err, false)
			}

			if err := j.sleep(ctx, backoff); err != nil {
				return err
			}

			backoff = time.Duration(float64(backoff) * j.config.BackoffFactor)

			if backoff > j.config.MaxBackoff && j.config.MaxBackoff > 0 {
				backoff = j.config.MaxBackoff
			}
		}
	}

	if j.tracker != nil {
		j.tracker.RecordFailure(j.Name(), lastErr, true)
	}

	_ = j.deadLetterHandler.HandleDeadLetter(ctx, j.Name(), lastErr, j.config.MaxRetries+1)
	return fmt.Errorf("job %s failed after %d attempts: %w", j.Name(), j.config.MaxRetries+1, lastErr)
}

package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

type failingJob struct {
	runs int
	err  error
}

func (j *failingJob) Name() string { return "failing-job" }
func (j *failingJob) Run(_ context.Context) error {
	j.runs++
	return j.err
}

type deadLetterTestHandler struct {
	deadLettered bool
	attempts     int
}

func (h *deadLetterTestHandler) HandleDeadLetter(_ context.Context, _ string, _ error, attempts int) error {
	h.deadLettered = true
	h.attempts = attempts
	return nil
}

func TestRetryableJobSuccess(t *testing.T) {
	tracker := NewStatusTracker()
	job := &failingJob{err: nil}
	retryable, err := NewRetryableJob(job, DefaultRetryConfig(), nil, tracker)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := retryable.Run(context.Background()); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if job.runs != 1 {
		t.Fatalf("expected 1 run, got %d", job.runs)
	}

	st, ok := tracker.Status("failing-job")

	if !ok || st.State != StateSucceeded {
		t.Fatalf("expected status succeeded, got %v", st)
	}
}

func TestRetryableJobFailureAndDeadLetter(t *testing.T) {
	tracker := NewStatusTracker()
	dlh := &deadLetterTestHandler{}
	job := &failingJob{err: errors.New("persistent error")}
	cfg := RetryConfig{MaxRetries: 2, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond, BackoffFactor: 1.5}
	retryable, err := NewRetryableJob(job, cfg, dlh, tracker)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := retryable.Run(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}

	if job.runs != 3 {
		t.Fatalf("expected 3 run attempts (1 initial + 2 retries), got %d", job.runs)
	}

	if !dlh.deadLettered || dlh.attempts != 3 {
		t.Fatalf("expected dead-letter callback with 3 attempts, got %v", dlh)
	}

	st, ok := tracker.Status("failing-job")

	if !ok || st.State != StateDeadLettered {
		t.Fatalf("expected status dead_lettered, got %v", st)
	}
}

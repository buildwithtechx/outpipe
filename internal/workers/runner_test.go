package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testJob struct {
	called bool
	err    error
}

func (j *testJob) Name() string { return "test" }

func (j *testJob) Run(context.Context) error {
	j.called = true
	return j.err
}

func TestRunnerRunsJobs(t *testing.T) {
	job := &testJob{}
	runner, err := NewRunner([]Job{job}, time.Second, nil)

	if err != nil {
		t.Fatal(err)
	}

	if err := runner.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !job.called {
		t.Fatal("expected job to run")
	}
}

func TestRunnerReturnsJobError(t *testing.T) {
	job := &testJob{err: errors.New("failed")}
	runner, err := NewRunner([]Job{job}, time.Second, nil)

	if err != nil {
		t.Fatal(err)
	}

	if err := runner.RunOnce(context.Background()); err == nil {
		t.Fatal("expected job error")
	}
}

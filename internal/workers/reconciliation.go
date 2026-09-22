package workers

import (
	"context"
	"fmt"
	"time"
)

type OperationalReconciler interface {
	Reconcile(context.Context, time.Time) error
}

type ReconciliationJob struct {
	reconciler OperationalReconciler
	now        func() time.Time
}

func NewReconciliationJob(reconciler OperationalReconciler) (*ReconciliationJob, error) {

	if reconciler == nil {
		return nil, fmt.Errorf("operational reconciler is required")
	}

	return &ReconciliationJob{reconciler: reconciler, now: time.Now}, nil
}

func (j *ReconciliationJob) Name() string {
	return "redis-operational-reconciliation"
}

func (j *ReconciliationJob) Run(ctx context.Context) error {

	if err := j.reconciler.Reconcile(ctx, j.now()); err != nil {
		return fmt.Errorf("reconcile redis operational state: %w", err)
	}

	return nil
}

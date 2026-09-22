package workers

import (
	"context"
	"fmt"
	"time"
)

type BillingReconciler interface {
	Reconcile(context.Context, time.Time) error
}

type BillingJob struct {
	reconciler BillingReconciler
	now        func() time.Time
}

func NewBillingJob(reconciler BillingReconciler) (*BillingJob, error) {

	if reconciler == nil {
		return nil, fmt.Errorf("billing reconciler is required")
	}

	return &BillingJob{reconciler: reconciler, now: time.Now}, nil
}

func (j *BillingJob) Name() string {
	return "billing-reconciliation"
}

func (j *BillingJob) Run(ctx context.Context) error {

	if err := j.reconciler.Reconcile(ctx, j.now()); err != nil {
		return fmt.Errorf("reconcile billing: %w", err)
	}

	return nil
}

package workers

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/services"
)

type MachineFleetReconciler interface {
	ReconcileAll(context.Context) (map[string]services.FleetReconcileResult, error)
}

type MachineFleetJob struct {
	reconciler MachineFleetReconciler
}

func NewMachineFleetJob(reconciler MachineFleetReconciler) (*MachineFleetJob, error) {
	if reconciler == nil {
		return nil, fmt.Errorf("machine fleet reconciler is required")
	}
	return &MachineFleetJob{reconciler: reconciler}, nil
}

func (j *MachineFleetJob) Name() string {
	return "machine-fleets"
}

func (j *MachineFleetJob) Run(ctx context.Context) error {
	if _, err := j.reconciler.ReconcileAll(ctx); err != nil {
		return fmt.Errorf("reconcile machine fleets: %w", err)
	}
	return nil
}

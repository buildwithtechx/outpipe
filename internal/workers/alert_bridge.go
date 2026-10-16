package workers

import (
	"context"
	"fmt"
)

type AlertBridgeProcessor interface {
	ProcessDueAlerts(context.Context, int) error
}

type AlertBridgeJob struct {
	processor AlertBridgeProcessor
}

func NewAlertBridgeJob(processor AlertBridgeProcessor) (*AlertBridgeJob, error) {
	if processor == nil {
		return nil, fmt.Errorf("alert bridge processor is required")
	}
	return &AlertBridgeJob{processor: processor}, nil
}

func (j *AlertBridgeJob) Name() string {
	return "alert-bridge"
}

func (j *AlertBridgeJob) Run(ctx context.Context) error {
	if err := j.processor.ProcessDueAlerts(ctx, 100); err != nil {
		return fmt.Errorf("process due alerts: %w", err)
	}
	return nil
}

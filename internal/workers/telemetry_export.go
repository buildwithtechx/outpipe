package workers

import (
	"context"
	"fmt"
)

type TelemetryExportProcessor interface {
	ProcessPendingExports(context.Context, int) error
}

type TelemetryExportJob struct {
	processor TelemetryExportProcessor
}

func NewTelemetryExportJob(processor TelemetryExportProcessor) (*TelemetryExportJob, error) {
	if processor == nil {
		return nil, fmt.Errorf("telemetry export processor is required")
	}
	return &TelemetryExportJob{processor: processor}, nil
}

func (j *TelemetryExportJob) Name() string {
	return "telemetry-export"
}

func (j *TelemetryExportJob) Run(ctx context.Context) error {
	if err := j.processor.ProcessPendingExports(ctx, 100); err != nil {
		return fmt.Errorf("process telemetry exports: %w", err)
	}
	return nil
}

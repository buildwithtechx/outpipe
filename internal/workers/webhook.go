package workers

import (
	"context"
	"fmt"
)

type WebhookDeliveryProcessor interface {
	ProcessPending(context.Context, int) error
}

type WebhookJob struct {
	processor WebhookDeliveryProcessor
}

func NewWebhookJob(processor WebhookDeliveryProcessor) (*WebhookJob, error) {
	if processor == nil {
		return nil, fmt.Errorf("webhook delivery processor is required")
	}
	return &WebhookJob{processor: processor}, nil
}

func (j *WebhookJob) Name() string {
	return "webhook-delivery"
}

func (j *WebhookJob) Run(ctx context.Context) error {
	if err := j.processor.ProcessPending(ctx, 100); err != nil {
		return fmt.Errorf("process webhook deliveries: %w", err)
	}
	return nil
}

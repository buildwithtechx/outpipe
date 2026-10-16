package workers

import (
	"context"
	"fmt"
)

type SubscriberNotificationProcessor interface {
	ProcessPendingSubscriberNotifications(context.Context, int) error
}

type SubscriberNotificationJob struct {
	processor SubscriberNotificationProcessor
}

func NewSubscriberNotificationJob(processor SubscriberNotificationProcessor) (*SubscriberNotificationJob, error) {
	if processor == nil {
		return nil, fmt.Errorf("subscriber notification processor is required")
	}
	return &SubscriberNotificationJob{processor: processor}, nil
}

func (j *SubscriberNotificationJob) Name() string {
	return "subscriber-notifications"
}

func (j *SubscriberNotificationJob) Run(ctx context.Context) error {
	if err := j.processor.ProcessPendingSubscriberNotifications(ctx, 100); err != nil {
		return fmt.Errorf("process subscriber notifications: %w", err)
	}
	return nil
}

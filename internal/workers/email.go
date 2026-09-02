package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"outpipe.dev/outpipe/internal/repositories"
)

type EmailSender interface {
	Send(context.Context, string, string, string) error
}

type EmailJob struct {
	deliveries repositories.EmailRepository
	sender     EmailSender
	logger     *slog.Logger
	now        func() time.Time
}

func NewEmailJob(deliveries repositories.EmailRepository, sender EmailSender, logger *slog.Logger) (*EmailJob, error) {
	if deliveries == nil || sender == nil {
		return nil, fmt.Errorf("email repository and sender are required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &EmailJob{deliveries: deliveries, sender: sender, logger: logger, now: time.Now}, nil
}

func (j *EmailJob) Name() string { return "email-delivery" }

func (j *EmailJob) Run(ctx context.Context) error {
	deliveries, err := j.deliveries.ClaimPending(ctx, j.now().UTC(), 50)
	if err != nil {
		return err
	}
	for _, delivery := range deliveries {
		if err := j.sender.Send(ctx, delivery.To, delivery.Subject, delivery.HTML); err != nil {
			next := j.now().UTC().Add(retryDelay(delivery.Attempts))
			if markErr := j.deliveries.MarkFailed(ctx, delivery.ID, err.Error(), next); markErr != nil {
				return fmt.Errorf("mark failed email %s: %w", delivery.ID, markErr)
			}
			j.logger.WarnContext(ctx, "email delivery failed", "delivery_id", delivery.ID, "error", err)
			continue
		}
		if err := j.deliveries.MarkSent(ctx, delivery.ID, j.now().UTC()); err != nil {
			return fmt.Errorf("mark sent email %s: %w", delivery.ID, err)
		}
	}
	return nil
}

func retryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	if attempts > 6 {
		attempts = 6
	}
	return time.Duration(1<<uint(attempts-1)) * time.Minute
}

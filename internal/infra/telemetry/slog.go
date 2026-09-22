package telemetry

import (
	"context"
	"log/slog"
)

type Event struct {
	Name       string
	Properties map[string]any
}

type Reporter interface {
	Report(context.Context, Event) error
}

type SlogReporter struct {
	logger *slog.Logger
}

func NewSlog(logger *slog.Logger) *SlogReporter {

	if logger == nil {
		logger = slog.Default()
	}

	return &SlogReporter{logger: logger}
}

func (r *SlogReporter) Report(_ context.Context, event Event) error {
	attributes := make([]any, 0, len(event.Properties)*2)

	for key, value := range event.Properties {
		attributes = append(attributes, slog.Any(key, value))
	}

	r.logger.Info("telemetry event", append([]any{slog.String("event", event.Name)}, attributes...)...)
	return nil
}

type NopReporter struct{}

func (NopReporter) Report(context.Context, Event) error {
	return nil
}

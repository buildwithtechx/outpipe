package services

import "fmt"

type TelemetryUnavailableError struct {
	Cause error `json:"-"`
}

func (e *TelemetryUnavailableError) Error() string {
	return fmt.Sprintf("telemetry backend unavailable: %v", e.Cause)
}
func (e *TelemetryUnavailableError) Unwrap() error { return e.Cause }

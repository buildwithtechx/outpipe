package services

import (
	"encoding/json"
	"fmt"
	"time"
)

func validateTelemetryTimestamp(values ...uint64) error {
	latest := time.Now().UTC().Add(5 * time.Minute)
	for _, value := range values {
		if value != 0 && telemetryTime(value).After(latest) {
			return fmt.Errorf("telemetry timestamp exceeds permitted clock skew")
		}
	}
	return nil
}

func analyticsEventTime(item map[string]json.RawMessage) time.Time {
	now := time.Now().UTC()
	for _, key := range []string{"timestamp", "start_time", "created_at"} {
		var raw string
		if err := json.Unmarshal(item[key], &raw); err != nil {
			continue
		}
		value, err := time.Parse(time.RFC3339Nano, raw)
		if err == nil && !value.Before(time.Unix(0, 0)) && !value.After(now.Add(5*time.Minute)) {
			return value.UTC()
		}
	}
	return now
}

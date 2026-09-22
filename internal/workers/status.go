package workers

import (
	"sync"
	"time"
)

type JobState string

const (
	StateIdle         JobState = "idle"
	StateRunning      JobState = "running"
	StateSucceeded    JobState = "succeeded"
	StateFailed       JobState = "failed"
	StateDeadLettered JobState = "dead_lettered"
)

type JobStatus struct {
	Name          string    `json:"name"`
	State         JobState  `json:"state"`
	RunCount      int64     `json:"runCount"`
	FailureCount  int64     `json:"failureCount"`
	LastRunAt     time.Time `json:"lastRunAt,omitempty"`
	LastSuccessAt time.Time `json:"lastSuccessAt,omitempty"`
	LastError     string    `json:"lastError,omitempty"`
}

type StatusTracker struct {
	mu       sync.RWMutex
	statuses map[string]JobStatus
}

func NewStatusTracker() *StatusTracker {
	return &StatusTracker{statuses: make(map[string]JobStatus)}
}

func (t *StatusTracker) RecordStart(name string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	st := t.statuses[name]
	st.Name = name
	st.State = StateRunning
	st.RunCount++
	st.LastRunAt = now
	t.statuses[name] = st
}

func (t *StatusTracker) RecordSuccess(name string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	st := t.statuses[name]
	st.Name = name
	st.State = StateSucceeded
	st.LastSuccessAt = now
	st.LastError = ""
	t.statuses[name] = st
}

func (t *StatusTracker) RecordFailure(name string, err error, deadLetter bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	st := t.statuses[name]
	st.Name = name

	if deadLetter {
		st.State = StateDeadLettered
	} else {
		st.State = StateFailed
	}

	st.FailureCount++

	if err != nil {
		st.LastError = err.Error()
	}

	t.statuses[name] = st
}

func (t *StatusTracker) Statuses() []JobStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]JobStatus, 0, len(t.statuses))

	for _, st := range t.statuses {
		result = append(result, st)
	}

	return result
}

func (t *StatusTracker) Status(name string) (JobStatus, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	st, ok := t.statuses[name]
	return st, ok
}

package main

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/engine"
	"time"
)

type capturePolicy struct {
	enabled bool
	expires time.Time
	pending bool
	orgID   string
}
type capturePolicyRequest struct {
	tunnelID string
	orgID    string
}

func (r *captureRecorder) start(ctx context.Context) {
	r.policies = make(map[string]capturePolicy)
	r.queue = make(chan engine.RequestCapture, 128)
	r.policyQueue = make(chan capturePolicyRequest, 128)
	for range 4 {
		r.wait.Add(2)
		go func() {
			defer r.wait.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case request := <-r.policyQueue:
					lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
					enabled, err := r.enabledPolicy(lookupCtx, request.tunnelID, request.orgID)
					cancel()
					r.mu.Lock()
					r.policies[request.tunnelID] = capturePolicy{enabled: err == nil && enabled, expires: time.Now().Add(time.Second), orgID: request.orgID}
					r.mu.Unlock()
				}
			}
		}()
		go func() {
			defer r.wait.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case capture := <-r.queue:
					for attempt := 0; attempt < 3 && ctx.Err() == nil; attempt++ {
						if err := r.sendCapture(ctx, capture); err == nil {
							break
						}
						if attempt == 2 {
							r.dropped.Add(1)
							break
						}
						timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
						select {
						case <-ctx.Done():
							timer.Stop()
							return
						case <-timer.C:
						}
					}
				}
			}
		}()
	}
}

func (r *captureRecorder) Enabled(_ context.Context, tunnelID, orgID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	policy := r.policies[tunnelID]
	if policy.expires.After(time.Now()) {
		return policy.orgID == orgID && policy.enabled, nil
	}
	if !policy.pending {
		if len(r.policies) >= 4096 {
			for key, entry := range r.policies {
				if !entry.pending && entry.expires.Before(time.Now()) {
					delete(r.policies, key)
				}
			}
			if len(r.policies) >= 4096 {
				return false, nil
			}
		}
		select {
		case r.policyQueue <- capturePolicyRequest{tunnelID: tunnelID, orgID: orgID}:
			policy.pending = true
			r.policies[tunnelID] = policy
		default:
		}
	}
	return false, nil
}

func (r *captureRecorder) RecordCapture(_ context.Context, capture engine.RequestCapture) error {
	select {
	case r.queue <- capture:
	default:
		r.dropped.Add(1)
		return fmt.Errorf("capture upload queue is full")
	}
	return nil
}

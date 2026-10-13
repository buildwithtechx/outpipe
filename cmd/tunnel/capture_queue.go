package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/engine"
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
type captureUpload struct {
	capture engine.RequestCapture
	spool   string
}

func (r *captureRecorder) start(ctx context.Context) {
	r.policies = make(map[string]capturePolicy)
	r.queue = make(chan captureUpload, 128)
	r.policyQueue = make(chan capturePolicyRequest, 128)
	interval := r.sweepInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	sweep := time.NewTicker(interval)
	go func() {
		<-ctx.Done()
		sweep.Stop()
	}()
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
				case upload := <-r.queue:
					r.deliverUpload(ctx, upload)
				case <-sweep.C:
					r.drainSpool(ctx)
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
	if capture.ID == "" {
		capture.ID = uuid.NewString()
	}
	if r.spool != nil {
		file, err := r.spool.store(capture)
		if err != nil {
			return r.recordMemory(capture)
		}
		select {
		case r.queue <- captureUpload{capture: capture, spool: file}:
		default:
		}
		return nil
	}
	return r.recordMemory(capture)
}

func (r *captureRecorder) recordMemory(capture engine.RequestCapture) error {
	select {
	case r.queue <- captureUpload{capture: capture}:
	default:
		r.dropped.Add(1)
		return fmt.Errorf("capture upload queue is full")
	}
	return nil
}

func (r *captureRecorder) uploadWithRetry(ctx context.Context, capture engine.RequestCapture) bool {
	for attempt := range 3 {
		if ctx.Err() != nil {
			return false
		}
		if err := r.sendCapture(ctx, capture); err == nil {
			return true
		}
		if attempt < 2 {
			timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return false
			case <-timer.C:
			}
		}
	}
	return false
}

func (r *captureRecorder) deliverUpload(ctx context.Context, upload captureUpload) {
	claimed := ""
	if upload.spool != "" && r.spool != nil {
		path, err := r.spool.claimPath(upload.spool)
		if err != nil {
			return
		}
		claimed = path
	}
	if r.uploadWithRetry(ctx, upload.capture) {
		if claimed != "" && r.spool != nil {
			r.spool.complete(claimed)
		}
		return
	}
	if ctx.Err() != nil {
		if claimed != "" && r.spool != nil {
			r.spool.release(claimed, spoolAttemptOfPath(claimed), time.Now())
		}
		return
	}
	if claimed != "" && r.spool != nil {
		attempt := spoolAttemptOfPath(claimed) + 3
		r.spool.release(claimed, attempt, time.Now().Add(r.spool.retryDelay(attempt)))
		return
	}
	r.dropped.Add(1)
}

func (r *captureRecorder) drainSpool(ctx context.Context) {
	if r.spool == nil || ctx.Err() != nil {
		return
	}
	now := time.Now()
	if evicted := r.spool.evict(now); evicted > 0 {
		r.dropped.Add(uint64(evicted))
	}
	for _, name := range r.spool.pending() {
		if ctx.Err() != nil {
			return
		}
		if strings.HasSuffix(name, ".sending") {
			continue
		}
		claimed, err := r.spool.claim(name)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(claimed)
		if err != nil {
			continue
		}
		var capture engine.RequestCapture
		if err := json.Unmarshal(data, &capture); err != nil {
			r.spool.complete(claimed)
			r.dropped.Add(1)
			continue
		}
		if r.uploadWithRetry(ctx, capture) {
			r.spool.complete(claimed)
			continue
		}
		if ctx.Err() != nil {
			r.spool.release(claimed, spoolAttempt(name), now)
			return
		}
		attempt := spoolAttempt(name) + 3
		r.spool.release(claimed, attempt, now.Add(r.spool.retryDelay(attempt)))
	}
}

package relay

import (
	"context"
	"fmt"
	"time"
)

type cachedTunnelPolicy struct {
	policy  ManagedTunnelPolicy
	expires time.Time
}

func (r *InternalTunnelResolver) cachedPolicy(ctx context.Context, id string) (ManagedTunnelPolicy, error) {
	r.mu.Lock()
	cached, ok := r.policies[id]
	r.mu.Unlock()
	if ok && time.Now().Before(cached.expires) {
		return cached.policy, nil
	}
	value, err, _ := r.requests.Do(id, func() (any, error) {
		lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		policy, err := r.resolve(lookupCtx, id)
		if err != nil {
			return nil, err
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if len(r.policies) >= 4096 {
			for key, item := range r.policies {
				if time.Now().After(item.expires) {
					delete(r.policies, key)
				}
			}
		}
		if len(r.policies) < 4096 {
			r.policies[id] = cachedTunnelPolicy{policy: policy, expires: time.Now().Add(time.Second)}
		}
		return policy, nil
	})
	if err != nil {
		return ManagedTunnelPolicy{}, fmt.Errorf("load managed policy: %w", err)
	}
	return value.(ManagedTunnelPolicy), nil
}

package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

type Session struct {
	ID             string
	OrganizationID string
	TunnelID       string
	PasswordHash   string
	BandwidthLimit int64
	ConnectedAt    time.Time
	LastActiveAt   time.Time
	Send           func(context.Context, protocol.Envelope) error
	Close          func()
}

type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]Session
	aliases  map[string]string
}

func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{sessions: make(map[string]Session), aliases: make(map[string]string)}
}

func (r *SessionRegistry) Reserve(session Session, takeover bool) error {
	return r.ReserveWithDrain(session, takeover, 0)
}

func (r *SessionRegistry) ReserveWithDrain(session Session, takeover bool, drain time.Duration) error {
	if session.ID == "" || session.TunnelID == "" {
		return fmt.Errorf("session and tunnel ids are required")
	}
	if session.Send == nil {
		return fmt.Errorf("session sender is required")
	}
	now := time.Now().UTC()
	if session.ConnectedAt.IsZero() {
		session.ConnectedAt = now
	}
	session.LastActiveAt = now

	r.mu.Lock()
	previous, exists := r.sessions[session.TunnelID]
	if exists && !takeover {
		r.mu.Unlock()
		return fmt.Errorf("tunnel %q is already connected", session.TunnelID)
	}
	r.sessions[session.TunnelID] = session
	r.mu.Unlock()
	if exists && previous.Close != nil {
		if drain <= 0 {
			previous.Close()
		} else {
			go func() {
				timer := time.NewTimer(drain)
				defer timer.Stop()
				<-timer.C
				previous.Close()
			}()
		}
	}
	return nil
}

func (r *SessionRegistry) Get(tunnelID string) (Session, bool) {
	r.mu.RLock()
	session, ok := r.sessions[tunnelID]
	r.mu.RUnlock()
	return session, ok
}

func (r *SessionRegistry) Touch(tunnelID string) bool {
	r.mu.Lock()
	session, ok := r.sessions[tunnelID]
	if ok {
		session.LastActiveAt = time.Now().UTC()
		r.sessions[tunnelID] = session
	}
	r.mu.Unlock()
	return ok
}

func (r *SessionRegistry) Remove(tunnelID string, sessionID string) bool {
	r.mu.Lock()
	session, ok := r.sessions[tunnelID]
	if !ok || (sessionID != "" && session.ID != sessionID) {
		r.mu.Unlock()
		return false
	}
	delete(r.sessions, tunnelID)
	for alias, target := range r.aliases {
		if target == tunnelID {
			delete(r.aliases, alias)
		}
	}
	r.mu.Unlock()
	return true
}

func (r *SessionRegistry) BindAlias(alias, tunnelID string) error {
	alias = normalizeAlias(alias)
	if alias == "" || tunnelID == "" {
		return fmt.Errorf("alias and tunnel id are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[tunnelID]; !ok {
		return fmt.Errorf("tunnel %q is not connected", tunnelID)
	}
	if previous, ok := r.aliases[alias]; ok && previous != tunnelID {
		return fmt.Errorf("route alias %q is already assigned", alias)
	}
	r.aliases[alias] = tunnelID
	return nil
}

func (r *SessionRegistry) Resolve(route string) (string, bool) {
	route = normalizeAlias(route)
	r.mu.RLock()
	tunnelID, ok := r.aliases[route]
	if !ok {
		_, ok = r.sessions[route]
		tunnelID = route
	}
	r.mu.RUnlock()
	return tunnelID, ok
}

func (r *SessionRegistry) Snapshot() []Session {
	r.mu.RLock()
	result := make([]Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		result = append(result, session)
	}
	r.mu.RUnlock()
	return result
}

func normalizeAlias(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

package sentinel

import (
	"sync"

	"github.com/langoai/lango/internal/eventbus"
)

// RevokeSessionFunc revokes all active session keys.
type RevokeSessionFunc func() error

// RestrictSessionFunc reduces session limits.
type RestrictSessionFunc func(factor float64) error

// SentinelAlertEvent wraps an Alert for the event bus.
type SentinelAlertEvent struct {
	Alert Alert
}

// EventName implements eventbus.Event.
func (e SentinelAlertEvent) EventName() string { return "sentinel.alert" }

// SessionGuard monitors sentinel alerts and manages session key safety.
type SessionGuard struct {
	bus        *eventbus.Bus
	revokeFn   RevokeSessionFunc
	restrictFn RestrictSessionFunc
	mu         sync.Mutex
	active     bool
}

// NewSessionGuard creates a session guard.
func NewSessionGuard(bus *eventbus.Bus) *SessionGuard {
	return &SessionGuard{bus: bus}
}

// SetRevokeFunc sets the callback for revoking sessions.
func (g *SessionGuard) SetRevokeFunc(fn RevokeSessionFunc) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.revokeFn = fn
}

// SetRestrictFunc sets the callback for restricting sessions.
func (g *SessionGuard) SetRestrictFunc(fn RestrictSessionFunc) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.restrictFn = fn
}

// Start subscribes to sentinel alert events.
func (g *SessionGuard) Start() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active {
		return
	}

	g.bus.Subscribe("sentinel.alert", func(ev eventbus.Event) {
		g.handleAlert(ev)
	})
	g.active = true
}

// handleAlert processes a sentinel alert and takes action.
func (g *SessionGuard) handleAlert(ev eventbus.Event) {
	alert, ok := ev.(SentinelAlertEvent)
	if !ok {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	switch alert.Alert.Severity {
	case SeverityCritical, SeverityHigh:
		if g.revokeFn != nil {
			_ = g.revokeFn()
		}
	case SeverityMedium:
		if g.restrictFn != nil {
			_ = g.restrictFn(0.5) // reduce limits by 50%
		}
	}
}

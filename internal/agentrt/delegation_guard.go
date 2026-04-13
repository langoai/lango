package agentrt

import (
	"sync"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

// CircuitState represents the circuit breaker state per agent.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota + 1
	CircuitOpen
	CircuitHalfOpen
)

type circuitBreaker struct {
	state         CircuitState
	failureCount  int
	lastFailureAt time.Time
	openedAt      time.Time
}

// DelegationGuard observes delegation events and maintains per-agent circuit breaker state.
// It does NOT make routing decisions — routing authority remains with the root orchestrator LLM.
type DelegationGuard struct {
	mu       sync.Mutex
	breakers map[string]*circuitBreaker
	cfg      config.CircuitBreakerCfg
	bus      *eventbus.Bus
}

// NewDelegationGuard creates a delegation guard.
func NewDelegationGuard(cfg config.CircuitBreakerCfg, bus *eventbus.Bus) *DelegationGuard {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = 30 * time.Second
	}
	return &DelegationGuard{
		breakers: make(map[string]*circuitBreaker),
		cfg:      cfg,
		bus:      bus,
	}
}

// IsOpen returns true if the agent's circuit is open (failing, should not receive delegations).
func (g *DelegationGuard) IsOpen(agentName string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	cb, ok := g.breakers[agentName]
	if !ok {
		return false
	}

	switch cb.state {
	case CircuitOpen:
		if time.Since(cb.openedAt) >= g.cfg.ResetTimeout {
			cb.state = CircuitHalfOpen
			return false
		}
		return true
	default:
		return false
	}
}

// RecordOutcome records the result of a delegation to update circuit breaker state.
func (g *DelegationGuard) RecordOutcome(agentName string, success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	cb := g.ensureBreaker(agentName)

	if success {
		cb.failureCount = 0
		cb.state = CircuitClosed
		return
	}

	cb.failureCount++
	cb.lastFailureAt = time.Now()

	if cb.failureCount >= g.cfg.FailureThreshold && cb.state != CircuitOpen {
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
		if g.bus != nil {
			g.bus.Publish(CircuitBreakerTrippedEvent{
				AgentName:    agentName,
				FailureCount: cb.failureCount,
				ResetAt:      cb.openedAt.Add(g.cfg.ResetTimeout),
			})
		}
	}
}

// State returns the current circuit state for an agent.
func (g *DelegationGuard) State(agentName string) CircuitState {
	g.mu.Lock()
	defer g.mu.Unlock()

	cb, ok := g.breakers[agentName]
	if !ok {
		return CircuitClosed
	}

	if cb.state == CircuitOpen && time.Since(cb.openedAt) >= g.cfg.ResetTimeout {
		cb.state = CircuitHalfOpen
	}
	return cb.state
}

// providerKey returns a namespaced key for provider-level circuit breakers
// to avoid collision with agent names.
func providerKey(provider string) string {
	return "provider:" + provider
}

// RecordProviderFailure records the result of a provider-level operation
// to update circuit breaker state. Provider keys are prefixed with "provider:"
// to avoid collision with agent names.
func (g *DelegationGuard) RecordProviderFailure(provider string, success bool) {
	g.RecordOutcome(providerKey(provider), success)
}

// IsProviderOpen returns true if the provider's circuit is open.
func (g *DelegationGuard) IsProviderOpen(provider string) bool {
	return g.IsOpen(providerKey(provider))
}

func (g *DelegationGuard) ensureBreaker(name string) *circuitBreaker {
	cb, ok := g.breakers[name]
	if !ok {
		cb = &circuitBreaker{state: CircuitClosed}
		g.breakers[name] = cb
	}
	return cb
}

package cockpit

import (
	"sync"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/eventbus"
)

// tokenSnapshot holds accumulated token usage for a single turn.
type tokenSnapshot struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	CacheTokens  int64
}

// runtimeStatus is a point-in-time view of runtime state for the context panel.
type runtimeStatus struct {
	ActiveAgent     string
	DelegationCount int
	TurnTokens      int64
	IsRunning       bool
}

// RuntimeTracker aggregates runtime events from the EventBus for TUI display.
// It tracks per-turn token usage, delegation counts, and forwards recovery
// decisions as tea.Msg to the TUI program. It is safe for concurrent use.
type RuntimeTracker struct {
	mu              sync.RWMutex
	localSessionKey string
	turnTokens      tokenSnapshot
	delegationCount int
	activeAgent     string
	turnActive      bool // true while a local turn is running
	sender          msgSender
	bus             *eventbus.Bus
}

// NewRuntimeTracker creates a tracker and subscribes to runtime events.
// If bus is nil, the tracker still works for manual operations but receives
// no events.
func NewRuntimeTracker(bus *eventbus.Bus, sender msgSender, localSessionKey string) *RuntimeTracker {
	t := &RuntimeTracker{
		localSessionKey: localSessionKey,
		sender:          sender,
		bus:             bus,
	}
	if bus != nil {
		eventbus.SubscribeTyped(bus, func(e eventbus.TokenUsageEvent) {
			// Accept events with empty SessionKey (the production publisher
			// in wireModelAdapterTokenUsage does not set SessionKey) or
			// events that match the local cockpit session.
			if e.SessionKey != "" && e.SessionKey != t.localSessionKey {
				return
			}
			t.mu.Lock()
			defer t.mu.Unlock()
			// Only accumulate while a local turn is active, preventing
			// tokens from channel or background turns in the same process.
			if !t.turnActive {
				return
			}
			t.turnTokens.InputTokens += e.InputTokens
			t.turnTokens.OutputTokens += e.OutputTokens
			t.turnTokens.TotalTokens += e.TotalTokens
			t.turnTokens.CacheTokens += e.CacheTokens
		})
		eventbus.SubscribeTyped(bus, func(e agentrt.RecoveryDecisionEvent) {
			if e.SessionKey != t.localSessionKey {
				return
			}
			if t.sender != nil {
				t.sender.Send(chat.RecoveryMsg{
					CauseClass: e.CauseClass,
					Action:     e.Action,
					Attempt:    e.Attempt,
					Backoff:    e.Backoff,
				})
			}
		})
	}
	return t
}

// FlushTurnTokens returns the accumulated token snapshot and resets the
// counters to zero. Intended to be called once per turn completion.
func (t *RuntimeTracker) FlushTurnTokens() tokenSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	snap := t.turnTokens
	t.turnTokens = tokenSnapshot{}
	return snap
}

// StartTurn marks the beginning of a local turn.
// Must be called before the turn starts so that token events are accumulated
// and the context panel shows the Runtime section.
func (t *RuntimeTracker) StartTurn() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.turnActive = true
}

// RecordDelegation records an outward agent-to-agent delegation (not returns).
func (t *RuntimeTracker) RecordDelegation(to string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.delegationCount++
	t.activeAgent = to
}

// SetActiveAgent updates the active agent label without incrementing the
// delegation counter. Used for orchestrator return hops.
func (t *RuntimeTracker) SetActiveAgent(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.activeAgent = name
}

// ResetTurn clears delegation count, active agent, and turn-active flag.
// Token counters are NOT cleared here — use FlushTurnTokens for that.
func (t *RuntimeTracker) ResetTurn() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.delegationCount = 0
	t.activeAgent = ""
	t.turnActive = false
}

// Snapshot returns the current runtime status for the context panel.
func (t *RuntimeTracker) Snapshot() runtimeStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return runtimeStatus{
		ActiveAgent:     t.activeAgent,
		DelegationCount: t.delegationCount,
		TurnTokens:      t.turnTokens.TotalTokens,
		IsRunning:       t.turnActive,
	}
}


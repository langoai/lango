package agentrt

import (
	"sync"

	"github.com/langoai/lango/internal/config"
)

// BudgetAlert describes a budget threshold crossing.
type BudgetAlert struct {
	Resource   string // "turns" or "delegations"
	Used       int
	Limit      int
	Percentage float64
}

// BudgetPolicy mirrors the inner executor's turn and delegation counts observationally.
// It does NOT enforce limits — the inner executor's hardcoded limits remain authoritative.
// Mirroring uses the same counting semantics as agent.go:350:
// only events with function calls that are not delegations count as turns.
type BudgetPolicy struct {
	mu sync.Mutex

	baseTurns       int
	delegationLimit int
	alertThreshold  float64
	onAlert         func(BudgetAlert)

	turnCount       int
	delegationCount int
	uniqueAgents    map[string]struct{}
	turnAlerted     bool
	delegAlerted    bool
}

// NewBudgetPolicy creates a budget policy from config.
func NewBudgetPolicy(cfg config.BudgetCfg) *BudgetPolicy {
	baseTurns := cfg.ToolCallLimit
	if baseTurns <= 0 {
		baseTurns = 50
	}
	delegLimit := cfg.DelegationLimit
	if delegLimit <= 0 {
		delegLimit = 15
	}
	threshold := cfg.AlertThreshold
	if threshold <= 0 {
		threshold = 0.8
	}
	return &BudgetPolicy{
		baseTurns:       baseTurns,
		delegationLimit: delegLimit,
		alertThreshold:  threshold,
		uniqueAgents:    make(map[string]struct{}),
	}
}

// SetAlertHandler sets the callback for budget alerts.
func (b *BudgetPolicy) SetAlertHandler(fn func(BudgetAlert)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onAlert = fn
}

// Clone returns a new per-run tracker with the same immutable thresholds and alert handler.
// Mutable counters are intentionally reset so concurrent turns never share observational state.
func (b *BudgetPolicy) Clone() *BudgetPolicy {
	b.mu.Lock()
	defer b.mu.Unlock()
	return &BudgetPolicy{
		baseTurns:       b.baseTurns,
		delegationLimit: b.delegationLimit,
		alertThreshold:  b.alertThreshold,
		onAlert:         b.onAlert,
		uniqueAgents:    make(map[string]struct{}),
	}
}

// Reset clears all counters for a new turn.
func (b *BudgetPolicy) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.turnCount = 0
	b.delegationCount = 0
	b.uniqueAgents = make(map[string]struct{})
	b.turnAlerted = false
	b.delegAlerted = false
}

// RecordTurn increments the turn counter.
// Should only be called for function-call events that are not delegations,
// matching inner budget semantics (agent.go:350).
func (b *BudgetPolicy) RecordTurn() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.turnCount++
	b.checkAlert("turns", b.turnCount, b.baseTurns, &b.turnAlerted)
}

// RecordDelegation records a delegation to a target agent.
func (b *BudgetPolicy) RecordDelegation(target string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegationCount++
	if target != "" && target != "lango-orchestrator" {
		b.uniqueAgents[target] = struct{}{}
	}
	b.checkAlert("delegations", b.delegationCount, b.delegationLimit, &b.delegAlerted)
}

// TurnCount returns the current mirrored turn count.
func (b *BudgetPolicy) TurnCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.turnCount
}

// DelegationCount returns the current mirrored delegation count.
func (b *BudgetPolicy) DelegationCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegationCount
}

// UniqueAgentCount returns the number of distinct agents delegated to.
func (b *BudgetPolicy) UniqueAgentCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.uniqueAgents)
}

func (b *BudgetPolicy) checkAlert(resource string, used, limit int, alerted *bool) {
	if *alerted || b.onAlert == nil {
		return
	}
	pct := float64(used) / float64(limit)
	if pct >= b.alertThreshold {
		*alerted = true
		b.onAlert(BudgetAlert{
			Resource:   resource,
			Used:       used,
			Limit:      limit,
			Percentage: pct,
		})
	}
}

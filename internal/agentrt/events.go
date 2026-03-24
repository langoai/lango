package agentrt

import "time"

// DelegationObservedEvent is published when a delegation is observed by the guard.
type DelegationObservedEvent struct {
	From      string
	To        string
	IsOpen    bool // true if target agent's circuit is open
	SessionID string
}

func (e DelegationObservedEvent) EventName() string { return "agent.delegation.observed" }

// BudgetAlertEvent is published when a budget threshold is crossed.
type BudgetAlertEvent struct {
	Resource   string  // "turns" or "delegations"
	Used       int
	Limit      int
	Percentage float64
	SessionID  string
}

func (e BudgetAlertEvent) EventName() string { return "agent.budget.alert" }

// RecoveryEvent is published when a recovery action is taken.
type RecoveryEvent struct {
	Action    RecoveryAction
	AgentName string
	Error     string
	SessionID string
}

func (e RecoveryEvent) EventName() string { return "agent.recovery" }

// CircuitBreakerTrippedEvent is published when an agent's circuit breaker opens.
type CircuitBreakerTrippedEvent struct {
	AgentName    string
	FailureCount int
	ResetAt      time.Time
}

func (e CircuitBreakerTrippedEvent) EventName() string { return "agent.circuit_breaker.tripped" }

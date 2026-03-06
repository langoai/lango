package eventbus

import "math/big"

// BudgetAlertEvent is published when a task budget crosses a configured threshold.
type BudgetAlertEvent struct {
	TaskID    string
	Threshold float64 // the threshold percentage that was crossed (e.g. 0.5, 0.8)
}

// EventName implements Event.
func (e BudgetAlertEvent) EventName() string { return "budget.alert" }

// BudgetExhaustedEvent is published when a task budget is fully consumed.
type BudgetExhaustedEvent struct {
	TaskID     string
	TotalSpent *big.Int
}

// EventName implements Event.
func (e BudgetExhaustedEvent) EventName() string { return "budget.exhausted" }

// NegotiationStartedEvent is published when a negotiation session begins.
type NegotiationStartedEvent struct {
	SessionID    string
	InitiatorDID string
	ResponderDID string
	ToolName     string
}

// EventName implements Event.
func (e NegotiationStartedEvent) EventName() string { return "negotiation.started" }

// NegotiationCompletedEvent is published when negotiation terms are agreed.
type NegotiationCompletedEvent struct {
	SessionID    string
	InitiatorDID string
	ResponderDID string
	AgreedPrice  *big.Int
}

// EventName implements Event.
func (e NegotiationCompletedEvent) EventName() string { return "negotiation.completed" }

// NegotiationFailedEvent is published when a negotiation fails or expires.
type NegotiationFailedEvent struct {
	SessionID string
	Reason    string // "rejected", "expired", "cancelled"
}

// EventName implements Event.
func (e NegotiationFailedEvent) EventName() string { return "negotiation.failed" }

// EscrowCreatedEvent is published when an escrow is locked.
type EscrowCreatedEvent struct {
	EscrowID string
	PayerDID string
	PayeeDID string
	Amount   *big.Int
}

// EventName implements Event.
func (e EscrowCreatedEvent) EventName() string { return "escrow.created" }

// EscrowMilestoneEvent is published when an escrow milestone is completed.
type EscrowMilestoneEvent struct {
	EscrowID    string
	MilestoneID string
	Index       int
}

// EventName implements Event.
func (e EscrowMilestoneEvent) EventName() string { return "escrow.milestone" }

// EscrowReleasedEvent is published when escrow funds are released on-chain.
type EscrowReleasedEvent struct {
	EscrowID string
	Amount   *big.Int
}

// EventName implements Event.
func (e EscrowReleasedEvent) EventName() string { return "escrow.released" }

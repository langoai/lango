package eventbus

import (
	"math/big"
	"time"
)

// Event name constants for economy domain events.
const (
	EventBudgetAlert           = "budget.alert"
	EventBudgetExhausted       = "budget.exhausted"
	EventNegotiationStarted    = "negotiation.started"
	EventNegotiationCompleted  = "negotiation.completed"
	EventNegotiationFailed     = "negotiation.failed"
	EventEscrowCreated         = "escrow.created"
	EventEscrowMilestone       = "escrow.milestone"
	EventEscrowReleased        = "escrow.released"
	EventEscrowOnChainDeposit  = "escrow.onchain.deposit"
	EventEscrowOnChainWork     = "escrow.onchain.work"
	EventEscrowOnChainRelease  = "escrow.onchain.release"
	EventEscrowOnChainRefund   = "escrow.onchain.refund"
	EventEscrowOnChainDispute  = "escrow.onchain.dispute"
	EventEscrowOnChainResolved = "escrow.onchain.resolved"
	EventEscrowReorgDetected   = "escrow.reorg.detected"
	EventEscrowDangling        = "escrow.dangling"
)

// BudgetAlertEvent is published when a task budget crosses a configured threshold.
type BudgetAlertEvent struct {
	TaskID    string
	Threshold float64 // the threshold percentage that was crossed (e.g. 0.5, 0.8)
}

// EventName implements Event.
func (e BudgetAlertEvent) EventName() string { return EventBudgetAlert }

// BudgetExhaustedEvent is published when a task budget is fully consumed.
type BudgetExhaustedEvent struct {
	TaskID     string
	TotalSpent *big.Int
}

// EventName implements Event.
func (e BudgetExhaustedEvent) EventName() string { return EventBudgetExhausted }

// NegotiationStartedEvent is published when a negotiation session begins.
type NegotiationStartedEvent struct {
	SessionID    string
	InitiatorDID string
	ResponderDID string
	ToolName     string
}

// EventName implements Event.
func (e NegotiationStartedEvent) EventName() string { return EventNegotiationStarted }

// NegotiationCompletedEvent is published when negotiation terms are agreed.
type NegotiationCompletedEvent struct {
	SessionID    string
	InitiatorDID string
	ResponderDID string
	AgreedPrice  *big.Int
}

// EventName implements Event.
func (e NegotiationCompletedEvent) EventName() string { return EventNegotiationCompleted }

// NegotiationFailedEvent is published when a negotiation fails or expires.
type NegotiationFailedEvent struct {
	SessionID string
	Reason    string // "rejected", "expired", "cancelled"
}

// EventName implements Event.
func (e NegotiationFailedEvent) EventName() string { return EventNegotiationFailed }

// EscrowCreatedEvent is published when an escrow is locked.
type EscrowCreatedEvent struct {
	EscrowID string
	PayerDID string
	PayeeDID string
	Amount   *big.Int
}

// EventName implements Event.
func (e EscrowCreatedEvent) EventName() string { return EventEscrowCreated }

// EscrowMilestoneEvent is published when an escrow milestone is completed.
type EscrowMilestoneEvent struct {
	EscrowID    string
	MilestoneID string
	Index       int
}

// EventName implements Event.
func (e EscrowMilestoneEvent) EventName() string { return EventEscrowMilestone }

// EscrowReleasedEvent is published when escrow funds are released on-chain.
type EscrowReleasedEvent struct {
	EscrowID string
	Amount   *big.Int
}

// EventName implements Event.
func (e EscrowReleasedEvent) EventName() string { return EventEscrowReleased }

// --- On-chain escrow events ---

// EscrowOnChainDepositEvent is published when tokens are deposited into an on-chain escrow.
type EscrowOnChainDepositEvent struct {
	EscrowID string
	DealID   string
	Buyer    string
	Amount   *big.Int
	TxHash   string
}

// EventName implements Event.
func (e EscrowOnChainDepositEvent) EventName() string { return EventEscrowOnChainDeposit }

// EscrowOnChainWorkEvent is published when work proof is submitted on-chain.
type EscrowOnChainWorkEvent struct {
	EscrowID string
	DealID   string
	Seller   string
	WorkHash string
	TxHash   string
}

// EventName implements Event.
func (e EscrowOnChainWorkEvent) EventName() string { return EventEscrowOnChainWork }

// EscrowOnChainReleaseEvent is published when on-chain escrow funds are released.
type EscrowOnChainReleaseEvent struct {
	EscrowID string
	DealID   string
	Seller   string
	Amount   *big.Int
	TxHash   string
}

// EventName implements Event.
func (e EscrowOnChainReleaseEvent) EventName() string { return EventEscrowOnChainRelease }

// EscrowOnChainRefundEvent is published when on-chain escrow funds are refunded.
type EscrowOnChainRefundEvent struct {
	EscrowID string
	DealID   string
	Buyer    string
	Amount   *big.Int
	TxHash   string
}

// EventName implements Event.
func (e EscrowOnChainRefundEvent) EventName() string { return EventEscrowOnChainRefund }

// EscrowOnChainDisputeEvent is published when an on-chain dispute is raised.
type EscrowOnChainDisputeEvent struct {
	EscrowID  string
	DealID    string
	Initiator string
	TxHash    string
}

// EventName implements Event.
func (e EscrowOnChainDisputeEvent) EventName() string { return EventEscrowOnChainDispute }

// EscrowOnChainResolvedEvent is published when an on-chain dispute is resolved.
type EscrowOnChainResolvedEvent struct {
	EscrowID    string
	DealID      string
	SellerFavor bool
	Amount      *big.Int
	TxHash      string
}

// EventName implements Event.
func (e EscrowOnChainResolvedEvent) EventName() string { return EventEscrowOnChainResolved }

// EscrowReorgDetectedEvent is published when a chain reorganization is detected
// by the event monitor (safeBlock < lastBlock).
type EscrowReorgDetectedEvent struct {
	PreviousBlock uint64
	NewBlock      uint64
	Depth         uint64
	ExceedsDepth  bool // reorg deeper than confirmationDepth
}

// EventName implements Event.
func (e EscrowReorgDetectedEvent) EventName() string { return EventEscrowReorgDetected }

// EscrowDanglingEvent is published when an escrow is stuck in Pending too long.
type EscrowDanglingEvent struct {
	EscrowID     string
	BuyerDID     string
	SellerDID    string
	Amount       string // string representation of *big.Int
	PendingSince time.Time
	Action       string // "expired", "refunded"
}

// EventName implements Event.
func (e EscrowDanglingEvent) EventName() string { return EventEscrowDangling }

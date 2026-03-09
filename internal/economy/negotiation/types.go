package negotiation

import (
	"math/big"
	"time"
)

// Phase represents the current phase of a negotiation.
type Phase string

const (
	PhaseProposed  Phase = "proposed"
	PhaseCountered Phase = "countered"
	PhaseAccepted  Phase = "accepted"
	PhaseRejected  Phase = "rejected"
	PhaseExpired   Phase = "expired"
	PhaseCancelled Phase = "cancelled"
)

// Terms represents the negotiated terms between two peers.
type Terms struct {
	Price      *big.Int      `json:"price"`
	Currency   string        `json:"currency"`
	ToolName   string        `json:"toolName"`
	MaxLatency time.Duration `json:"maxLatency,omitempty"`
	UseEscrow  bool          `json:"useEscrow"`
	EscrowID   string        `json:"escrowId,omitempty"`
}

// NegotiationSession tracks the state of a negotiation between two peers.
type NegotiationSession struct {
	ID           string     `json:"id"`
	InitiatorDID string     `json:"initiatorDid"`
	ResponderDID string     `json:"responderDid"`
	Phase        Phase      `json:"phase"`
	CurrentTerms *Terms     `json:"currentTerms"`
	Proposals    []Proposal `json:"proposals"`
	Round        int        `json:"round"`
	MaxRounds    int        `json:"maxRounds"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	ExpiresAt    time.Time  `json:"expiresAt"`
}

// IsTerminal returns true if the negotiation has reached a final state.
func (ns *NegotiationSession) IsTerminal() bool {
	switch ns.Phase {
	case PhaseAccepted, PhaseRejected, PhaseExpired, PhaseCancelled:
		return true
	}
	return false
}

// CanCounter returns true if the current round allows another counter.
func (ns *NegotiationSession) CanCounter() bool {
	return !ns.IsTerminal() && ns.Round < ns.MaxRounds
}

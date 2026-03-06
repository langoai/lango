package escrow

import (
	"math/big"
	"time"
)

// EscrowStatus represents the current state of an escrow.
type EscrowStatus string

const (
	StatusPending   EscrowStatus = "pending"
	StatusFunded    EscrowStatus = "funded"
	StatusActive    EscrowStatus = "active"
	StatusCompleted EscrowStatus = "completed"
	StatusReleased  EscrowStatus = "released"
	StatusDisputed  EscrowStatus = "disputed"
	StatusExpired   EscrowStatus = "expired"
	StatusRefunded  EscrowStatus = "refunded"
)

// MilestoneStatus represents the status of a single milestone.
type MilestoneStatus string

const (
	MilestonePending   MilestoneStatus = "pending"
	MilestoneCompleted MilestoneStatus = "completed"
	MilestoneDisputed  MilestoneStatus = "disputed"
)

// Milestone represents a deliverable checkpoint within an escrow.
type Milestone struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Amount      *big.Int        `json:"amount"`
	Status      MilestoneStatus `json:"status"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	Evidence    string          `json:"evidence,omitempty"`
}

// EscrowEntry represents a single escrow agreement between two peers.
type EscrowEntry struct {
	ID          string       `json:"id"`
	BuyerDID    string       `json:"buyerDid"`
	SellerDID   string       `json:"sellerDid"`
	TotalAmount *big.Int     `json:"totalAmount"`
	Status      EscrowStatus `json:"status"`
	Milestones  []Milestone  `json:"milestones"`
	TaskID      string       `json:"taskId,omitempty"`
	Reason      string       `json:"reason"`
	DisputeNote string       `json:"disputeNote,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	ExpiresAt   time.Time    `json:"expiresAt"`
}

// CompletedMilestones returns the count of completed milestones.
func (e *EscrowEntry) CompletedMilestones() int {
	count := 0
	for _, m := range e.Milestones {
		if m.Status == MilestoneCompleted {
			count++
		}
	}
	return count
}

// AllMilestonesCompleted returns true if every milestone is completed.
func (e *EscrowEntry) AllMilestonesCompleted() bool {
	if len(e.Milestones) == 0 {
		return false
	}
	return e.CompletedMilestones() == len(e.Milestones)
}

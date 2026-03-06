package escrow

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrMilestoneNotFound = errors.New("milestone not found")
	ErrEscrowExpired     = errors.New("escrow expired")
	ErrNoMilestones      = errors.New("escrow has no milestones")
	ErrTooManyMilestones = errors.New("too many milestones")
	ErrInvalidAmount     = errors.New("milestone amounts do not match total")
)

// SettlementExecutor handles actual fund transfer operations.
type SettlementExecutor interface {
	Lock(ctx context.Context, buyerDID string, amount *big.Int) error
	Release(ctx context.Context, sellerDID string, amount *big.Int) error
	Refund(ctx context.Context, buyerDID string, amount *big.Int) error
}

// EngineConfig holds engine configuration.
type EngineConfig struct {
	DefaultTimeout time.Duration
	MaxMilestones  int
	AutoRelease    bool
	DisputeWindow  time.Duration
}

// DefaultEngineConfig returns sensible defaults.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		DefaultTimeout: 24 * time.Hour,
		MaxMilestones:  10,
		AutoRelease:    false,
		DisputeWindow:  1 * time.Hour,
	}
}

// Engine manages the escrow lifecycle.
type Engine struct {
	store    Store
	settler  SettlementExecutor
	cfg      EngineConfig
	mu       sync.Mutex
	nowFunc  func() time.Time
}

// NewEngine creates a new escrow engine.
func NewEngine(store Store, settler SettlementExecutor, cfg EngineConfig) *Engine {
	return &Engine{
		store:   store,
		settler: settler,
		cfg:     cfg,
		nowFunc: time.Now,
	}
}

// CreateRequest holds the parameters for creating an escrow.
type CreateRequest struct {
	BuyerDID  string
	SellerDID string
	Amount    *big.Int
	Reason    string
	TaskID    string
	Milestones []MilestoneRequest
	ExpiresAt *time.Time
}

// MilestoneRequest defines a milestone at creation time.
type MilestoneRequest struct {
	Description string
	Amount      *big.Int
}

// Create initializes a new escrow in pending state.
func (e *Engine) Create(ctx context.Context, req CreateRequest) (*EscrowEntry, error) {
	if len(req.Milestones) == 0 {
		return nil, ErrNoMilestones
	}
	if e.cfg.MaxMilestones > 0 && len(req.Milestones) > e.cfg.MaxMilestones {
		return nil, fmt.Errorf("got %d milestones (max %d): %w", len(req.Milestones), e.cfg.MaxMilestones, ErrTooManyMilestones)
	}

	total := new(big.Int)
	milestones := make([]Milestone, len(req.Milestones))
	for i, mr := range req.Milestones {
		total.Add(total, mr.Amount)
		milestones[i] = Milestone{
			ID:          uuid.New().String(),
			Description: mr.Description,
			Amount:      new(big.Int).Set(mr.Amount),
			Status:      MilestonePending,
		}
	}

	if total.Cmp(req.Amount) != 0 {
		return nil, fmt.Errorf("milestone sum %s != total %s: %w", total.String(), req.Amount.String(), ErrInvalidAmount)
	}

	expiresAt := e.nowFunc().Add(e.cfg.DefaultTimeout)
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	entry := &EscrowEntry{
		ID:          uuid.New().String(),
		BuyerDID:    req.BuyerDID,
		SellerDID:   req.SellerDID,
		TotalAmount: new(big.Int).Set(req.Amount),
		Status:      StatusPending,
		Milestones:  milestones,
		TaskID:      req.TaskID,
		Reason:      req.Reason,
		ExpiresAt:   expiresAt,
	}

	if err := e.store.Create(entry); err != nil {
		return nil, fmt.Errorf("create escrow: %w", err)
	}
	return entry, nil
}

// Fund locks funds and transitions pending -> funded.
func (e *Engine) Fund(ctx context.Context, escrowID string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := e.checkExpiry(entry); err != nil {
		return nil, err
	}

	if err := validateTransition(entry.Status, StatusFunded); err != nil {
		return nil, err
	}

	if err := e.settler.Lock(ctx, entry.BuyerDID, entry.TotalAmount); err != nil {
		return nil, fmt.Errorf("lock funds: %w", err)
	}

	entry.Status = StatusFunded
	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Activate transitions funded -> active (work begins).
func (e *Engine) Activate(ctx context.Context, escrowID string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := e.checkExpiry(entry); err != nil {
		return nil, err
	}

	if err := validateTransition(entry.Status, StatusActive); err != nil {
		return nil, err
	}

	entry.Status = StatusActive
	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// CompleteMilestone marks a specific milestone as completed.
func (e *Engine) CompleteMilestone(ctx context.Context, escrowID, milestoneID, evidence string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := e.checkExpiry(entry); err != nil {
		return nil, err
	}

	if entry.Status != StatusActive {
		return nil, fmt.Errorf("complete milestone on %q status: %w", entry.Status, ErrInvalidTransition)
	}

	idx := -1
	for i, m := range entry.Milestones {
		if m.ID == milestoneID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("milestone %q: %w", milestoneID, ErrMilestoneNotFound)
	}

	now := e.nowFunc()
	entry.Milestones[idx].Status = MilestoneCompleted
	entry.Milestones[idx].CompletedAt = &now
	entry.Milestones[idx].Evidence = evidence

	if entry.AllMilestonesCompleted() {
		entry.Status = StatusCompleted
		if e.cfg.AutoRelease {
			if err := e.settler.Release(ctx, entry.SellerDID, entry.TotalAmount); err != nil {
				return nil, fmt.Errorf("auto-release: %w", err)
			}
			entry.Status = StatusReleased
		}
	}

	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Release transfers funds to the seller. Only from completed (or active with all milestones done).
func (e *Engine) Release(ctx context.Context, escrowID string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(entry.Status, StatusReleased); err != nil {
		return nil, err
	}

	if err := e.settler.Release(ctx, entry.SellerDID, entry.TotalAmount); err != nil {
		return nil, fmt.Errorf("release funds: %w", err)
	}

	entry.Status = StatusReleased
	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Dispute transitions to disputed state.
func (e *Engine) Dispute(ctx context.Context, escrowID, note string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(entry.Status, StatusDisputed); err != nil {
		return nil, err
	}

	entry.Status = StatusDisputed
	entry.DisputeNote = note
	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Refund returns funds to the buyer from a disputed escrow.
func (e *Engine) Refund(ctx context.Context, escrowID string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(entry.Status, StatusRefunded); err != nil {
		return nil, err
	}

	if err := e.settler.Refund(ctx, entry.BuyerDID, entry.TotalAmount); err != nil {
		return nil, fmt.Errorf("refund: %w", err)
	}

	entry.Status = StatusRefunded
	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Expire marks a timed-out escrow as expired and refunds if funded.
func (e *Engine) Expire(ctx context.Context, escrowID string) (*EscrowEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, err := e.store.Get(escrowID)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(entry.Status, StatusExpired); err != nil {
		return nil, err
	}

	// Refund if funds were locked.
	if entry.Status == StatusFunded || entry.Status == StatusActive {
		if err := e.settler.Refund(ctx, entry.BuyerDID, entry.TotalAmount); err != nil {
			return nil, fmt.Errorf("expire refund: %w", err)
		}
	}

	entry.Status = StatusExpired
	if err := e.store.Update(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Get returns an escrow by ID.
func (e *Engine) Get(id string) (*EscrowEntry, error) {
	return e.store.Get(id)
}

// List returns all escrows.
func (e *Engine) List() []*EscrowEntry {
	return e.store.List()
}

// ListByPeer returns escrows involving a specific peer.
func (e *Engine) ListByPeer(peerDID string) []*EscrowEntry {
	return e.store.ListByPeer(peerDID)
}

// checkExpiry checks if an escrow has expired and transitions it if so.
func (e *Engine) checkExpiry(entry *EscrowEntry) error {
	if e.nowFunc().After(entry.ExpiresAt) && canTransition(entry.Status, StatusExpired) {
		entry.Status = StatusExpired
		_ = e.store.Update(entry)
		return fmt.Errorf("escrow %q: %w", entry.ID, ErrEscrowExpired)
	}
	return nil
}

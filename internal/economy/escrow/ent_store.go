package escrow

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/escrowdeal"
)

// Compile-time interface check.
var _ Store = (*EntStore)(nil)

// EntStore implements Store using ent ORM with persistent storage.
type EntStore struct {
	client *ent.Client
}

// NewEntStore creates a new ent-backed escrow store.
func NewEntStore(client *ent.Client) *EntStore {
	return &EntStore{client: client}
}

// Create persists a new escrow entry.
func (s *EntStore) Create(entry *EscrowEntry) error {
	ctx := context.Background()
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	milestoneData, err := json.Marshal(entry.Milestones)
	if err != nil {
		return fmt.Errorf("marshal milestones: %w", err)
	}

	builder := s.client.EscrowDeal.Create().
		SetEscrowID(entry.ID).
		SetBuyerDid(entry.BuyerDID).
		SetSellerDid(entry.SellerDID).
		SetTotalAmount(entry.TotalAmount.String()).
		SetStatus(string(entry.Status)).
		SetMilestones(milestoneData).
		SetReason(entry.Reason).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		SetExpiresAt(entry.ExpiresAt)

	if entry.TaskID != "" {
		builder.SetTaskID(entry.TaskID)
	}
	if entry.DisputeNote != "" {
		builder.SetDisputeNote(entry.DisputeNote)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("create %q: %w", entry.ID, ErrEscrowExists)
		}
		return fmt.Errorf("create %q: %w", entry.ID, err)
	}
	return nil
}

// Get retrieves an escrow entry by ID.
func (s *EntStore) Get(id string) (*EscrowEntry, error) {
	ctx := context.Background()

	deal, err := s.client.EscrowDeal.Query().
		Where(escrowdeal.EscrowID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("get %q: %w", id, ErrEscrowNotFound)
		}
		return nil, fmt.Errorf("get %q: %w", id, err)
	}
	return dealToEntry(deal)
}

// List returns all escrow entries.
func (s *EntStore) List() []*EscrowEntry {
	ctx := context.Background()

	deals, err := s.client.EscrowDeal.Query().
		Order(escrowdeal.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil
	}

	result := make([]*EscrowEntry, 0, len(deals))
	for _, d := range deals {
		entry, err := dealToEntry(d)
		if err != nil {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// ListByStatus returns escrow entries matching the given status.
func (s *EntStore) ListByStatus(status EscrowStatus) []*EscrowEntry {
	ctx := context.Background()

	deals, err := s.client.EscrowDeal.Query().
		Where(escrowdeal.Status(string(status))).
		Order(escrowdeal.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil
	}

	result := make([]*EscrowEntry, 0, len(deals))
	for _, d := range deals {
		entry, err := dealToEntry(d)
		if err != nil {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// ListByStatusBefore returns escrow entries matching the status created before the given time.
func (s *EntStore) ListByStatusBefore(status EscrowStatus, before time.Time) []*EscrowEntry {
	ctx := context.Background()

	deals, err := s.client.EscrowDeal.Query().
		Where(
			escrowdeal.Status(string(status)),
			escrowdeal.CreatedAtLT(before),
		).
		Order(escrowdeal.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil
	}

	result := make([]*EscrowEntry, 0, len(deals))
	for _, d := range deals {
		entry, err := dealToEntry(d)
		if err != nil {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// ListByPeer returns escrow entries where the peer is buyer or seller.
func (s *EntStore) ListByPeer(peerDID string) []*EscrowEntry {
	ctx := context.Background()

	deals, err := s.client.EscrowDeal.Query().
		Where(
			escrowdeal.Or(
				escrowdeal.BuyerDid(peerDID),
				escrowdeal.SellerDid(peerDID),
			),
		).
		Order(escrowdeal.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil
	}

	result := make([]*EscrowEntry, 0, len(deals))
	for _, d := range deals {
		entry, err := dealToEntry(d)
		if err != nil {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// Update updates an existing escrow entry.
func (s *EntStore) Update(entry *EscrowEntry) error {
	ctx := context.Background()

	deal, err := s.client.EscrowDeal.Query().
		Where(escrowdeal.EscrowID(entry.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("update %q: %w", entry.ID, ErrEscrowNotFound)
		}
		return fmt.Errorf("update %q: %w", entry.ID, err)
	}

	milestoneData, err := json.Marshal(entry.Milestones)
	if err != nil {
		return fmt.Errorf("marshal milestones: %w", err)
	}

	now := time.Now()
	entry.UpdatedAt = now

	builder := deal.Update().
		SetBuyerDid(entry.BuyerDID).
		SetSellerDid(entry.SellerDID).
		SetTotalAmount(entry.TotalAmount.String()).
		SetStatus(string(entry.Status)).
		SetMilestones(milestoneData).
		SetReason(entry.Reason).
		SetUpdatedAt(now).
		SetExpiresAt(entry.ExpiresAt)

	if entry.TaskID != "" {
		builder.SetTaskID(entry.TaskID)
	} else {
		builder.ClearTaskID()
	}
	if entry.DisputeNote != "" {
		builder.SetDisputeNote(entry.DisputeNote)
	} else {
		builder.ClearDisputeNote()
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update %q: %w", entry.ID, err)
	}
	return nil
}

// Delete removes an escrow entry by ID.
func (s *EntStore) Delete(id string) error {
	ctx := context.Background()

	n, err := s.client.EscrowDeal.Delete().
		Where(escrowdeal.EscrowID(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete %q: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("delete %q: %w", id, ErrEscrowNotFound)
	}
	return nil
}

// SetOnChainDealID sets the on-chain deal ID for an escrow.
func (s *EntStore) SetOnChainDealID(escrowID, dealID string) error {
	ctx := context.Background()

	n, err := s.client.EscrowDeal.Update().
		Where(escrowdeal.EscrowID(escrowID)).
		SetOnChainDealID(dealID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("set on-chain deal ID %q: %w", escrowID, err)
	}
	if n == 0 {
		return fmt.Errorf("set on-chain deal ID %q: %w", escrowID, ErrEscrowNotFound)
	}
	return nil
}

// GetByOnChainDealID retrieves an escrow entry by its on-chain deal ID.
func (s *EntStore) GetByOnChainDealID(dealID string) (*EscrowEntry, error) {
	ctx := context.Background()

	deal, err := s.client.EscrowDeal.Query().
		Where(escrowdeal.OnChainDealID(dealID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("get by on-chain deal ID %q: %w", dealID, ErrEscrowNotFound)
		}
		return nil, fmt.Errorf("get by on-chain deal ID %q: %w", dealID, err)
	}
	return dealToEntry(deal)
}

// SetTxHash sets a transaction hash field on an escrow entry.
// The field parameter must be one of: TxDeposit, TxRelease, TxRefund.
func (s *EntStore) SetTxHash(escrowID string, field TransactionType, txHash string) error {
	ctx := context.Background()

	deal, err := s.client.EscrowDeal.Query().
		Where(escrowdeal.EscrowID(escrowID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("set tx hash %q: %w", escrowID, ErrEscrowNotFound)
		}
		return fmt.Errorf("set tx hash %q: %w", escrowID, err)
	}

	builder := deal.Update()
	switch field {
	case TxDeposit:
		builder.SetDepositTxHash(txHash)
	case TxRelease:
		builder.SetReleaseTxHash(txHash)
	case TxRefund:
		builder.SetRefundTxHash(txHash)
	default:
		return fmt.Errorf("set tx hash %q: unknown field %q", escrowID, field)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("set tx hash %q: %w", escrowID, err)
	}
	return nil
}

// dealToEntry converts an ent EscrowDeal to a domain EscrowEntry.
func dealToEntry(d *ent.EscrowDeal) (*EscrowEntry, error) {
	amount := new(big.Int)
	if _, ok := amount.SetString(d.TotalAmount, 10); !ok {
		return nil, fmt.Errorf("parse total amount %q", d.TotalAmount)
	}

	var milestones []Milestone
	if len(d.Milestones) > 0 {
		if err := json.Unmarshal(d.Milestones, &milestones); err != nil {
			return nil, fmt.Errorf("unmarshal milestones: %w", err)
		}
	}

	return &EscrowEntry{
		ID:          d.EscrowID,
		BuyerDID:    d.BuyerDid,
		SellerDID:   d.SellerDid,
		TotalAmount: amount,
		Status:      EscrowStatus(d.Status),
		Milestones:  milestones,
		TaskID:      d.TaskID,
		Reason:      d.Reason,
		DisputeNote: d.DisputeNote,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		ExpiresAt:   d.ExpiresAt,
	}, nil
}

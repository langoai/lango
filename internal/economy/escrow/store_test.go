package escrow

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEntry(id, buyer, seller string) *EscrowEntry {
	return &EscrowEntry{
		ID:          id,
		BuyerDID:    buyer,
		SellerDID:   seller,
		TotalAmount: big.NewInt(1000),
		Status:      StatusPending,
		Milestones: []Milestone{
			{ID: "m1", Description: "first", Amount: big.NewInt(500), Status: MilestonePending},
			{ID: "m2", Description: "second", Amount: big.NewInt(500), Status: MilestonePending},
		},
		Reason:    "test escrow",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func TestStoreCreate(t *testing.T) {
	tests := []struct {
		give    string
		setup   func(Store)
		entry   *EscrowEntry
		wantErr error
	}{
		{
			give:  "success",
			setup: func(s Store) {},
			entry: newTestEntry("e1", "did:buyer:1", "did:seller:1"),
		},
		{
			give: "duplicate ID",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
			},
			entry:   newTestEntry("e1", "did:buyer:2", "did:seller:2"),
			wantErr: ErrEscrowExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewMemoryStore()
			tt.setup(s)

			err := s.Create(tt.entry)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)

			got, err := s.Get(tt.entry.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.entry.ID, got.ID)
			assert.False(t, got.CreatedAt.IsZero())
			assert.False(t, got.UpdatedAt.IsZero())
		})
	}
}

func TestStoreGet(t *testing.T) {
	tests := []struct {
		give    string
		id      string
		setup   func(Store)
		wantErr error
	}{
		{
			give: "found",
			id:   "e1",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
			},
		},
		{
			give:    "not found",
			id:      "missing",
			setup:   func(s Store) {},
			wantErr: ErrEscrowNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewMemoryStore()
			tt.setup(s)

			got, err := s.Get(tt.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ID)
		})
	}
}

func TestStoreList(t *testing.T) {
	tests := []struct {
		give     string
		setup    func(Store)
		wantLen  int
	}{
		{
			give:    "empty store",
			setup:   func(s Store) {},
			wantLen: 0,
		},
		{
			give: "multiple entries",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
				_ = s.Create(newTestEntry("e2", "did:buyer:2", "did:seller:2"))
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewMemoryStore()
			tt.setup(s)
			assert.Len(t, s.List(), tt.wantLen)
		})
	}
}

func TestStoreListByPeer(t *testing.T) {
	tests := []struct {
		give    string
		peerDID string
		setup   func(Store)
		wantLen int
	}{
		{
			give:    "no matches",
			peerDID: "did:nobody",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
			},
			wantLen: 0,
		},
		{
			give:    "matches as buyer",
			peerDID: "did:buyer:1",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
				_ = s.Create(newTestEntry("e2", "did:buyer:2", "did:seller:2"))
			},
			wantLen: 1,
		},
		{
			give:    "matches as seller",
			peerDID: "did:seller:1",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
			},
			wantLen: 1,
		},
		{
			give:    "matches both roles",
			peerDID: "did:peer:1",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:peer:1", "did:seller:1"))
				_ = s.Create(newTestEntry("e2", "did:buyer:2", "did:peer:1"))
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewMemoryStore()
			tt.setup(s)
			assert.Len(t, s.ListByPeer(tt.peerDID), tt.wantLen)
		})
	}
}

func TestStoreUpdate(t *testing.T) {
	tests := []struct {
		give    string
		setup   func(Store)
		entry   *EscrowEntry
		wantErr error
	}{
		{
			give: "success",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
			},
			entry: &EscrowEntry{
				ID:          "e1",
				BuyerDID:    "did:buyer:1",
				SellerDID:   "did:seller:1",
				TotalAmount: big.NewInt(2000),
				Status:      StatusFunded,
			},
		},
		{
			give:  "not found",
			setup: func(s Store) {},
			entry: &EscrowEntry{
				ID:          "missing",
				TotalAmount: big.NewInt(100),
			},
			wantErr: ErrEscrowNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewMemoryStore()
			tt.setup(s)

			err := s.Update(tt.entry)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)

			got, err := s.Get(tt.entry.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.entry.Status, got.Status)
			assert.False(t, got.UpdatedAt.IsZero())
		})
	}
}

func TestStoreDelete(t *testing.T) {
	tests := []struct {
		give    string
		id      string
		setup   func(Store)
		wantErr error
	}{
		{
			give: "success",
			id:   "e1",
			setup: func(s Store) {
				_ = s.Create(newTestEntry("e1", "did:buyer:1", "did:seller:1"))
			},
		},
		{
			give:    "not found",
			id:      "missing",
			setup:   func(s Store) {},
			wantErr: ErrEscrowNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			s := NewMemoryStore()
			tt.setup(s)

			err := s.Delete(tt.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)

			_, err = s.Get(tt.id)
			assert.True(t, errors.Is(err, ErrEscrowNotFound))
		})
	}
}

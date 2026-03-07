package budget

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Allocate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		giveID    string
		giveTotal int64
		setup     func(*Store)
		wantErr   error
	}{
		{
			give:      "new budget succeeds",
			giveID:    "task-1",
			giveTotal: 1000000,
			wantErr:   nil,
		},
		{
			give:      "duplicate budget fails",
			giveID:    "task-1",
			giveTotal: 1000000,
			setup: func(s *Store) {
				_, _ = s.Allocate("task-1", big.NewInt(500000))
			},
			wantErr: ErrBudgetExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			got, err := s.Allocate(tt.giveID, big.NewInt(tt.giveTotal))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.giveID, got.TaskID)
			assert.Equal(t, 0, got.TotalBudget.Cmp(big.NewInt(tt.giveTotal)))
			assert.Equal(t, StatusActive, got.Status)
			assert.Equal(t, 0, got.Spent.Sign())
			assert.Equal(t, 0, got.Reserved.Sign())
		})
	}
}

func TestStore_Allocate_CopiesTotal(t *testing.T) {
	t.Parallel()

	s := NewStore()
	total := big.NewInt(1000000)
	tb, err := s.Allocate("task-1", total)
	require.NoError(t, err)

	total.SetInt64(0)
	assert.Equal(t, 0, tb.TotalBudget.Cmp(big.NewInt(1000000)))
}

func TestStore_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		giveID  string
		setup   func(*Store)
		wantErr error
	}{
		{
			give:   "existing budget",
			giveID: "task-1",
			setup: func(s *Store) {
				_, _ = s.Allocate("task-1", big.NewInt(1000000))
			},
			wantErr: nil,
		},
		{
			give:    "missing budget",
			giveID:  "task-999",
			wantErr: ErrBudgetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			got, err := s.Get(tt.giveID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.giveID, got.TaskID)
		})
	}
}

func TestStore_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		setup     func(*Store)
		wantCount int
	}{
		{
			give:      "empty store",
			wantCount: 0,
		},
		{
			give: "single budget",
			setup: func(s *Store) {
				_, _ = s.Allocate("task-1", big.NewInt(1000000))
			},
			wantCount: 1,
		},
		{
			give: "multiple budgets",
			setup: func(s *Store) {
				_, _ = s.Allocate("task-1", big.NewInt(1000000))
				_, _ = s.Allocate("task-2", big.NewInt(2000000))
				_, _ = s.Allocate("task-3", big.NewInt(3000000))
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			got := s.List()
			assert.Len(t, got, tt.wantCount)
		})
	}
}

func TestStore_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveID     string
		giveStatus BudgetStatus
		setup      func(*Store)
		wantErr    error
	}{
		{
			give:       "update existing budget",
			giveID:     "task-1",
			giveStatus: StatusExhausted,
			setup: func(s *Store) {
				_, _ = s.Allocate("task-1", big.NewInt(1000000))
			},
			wantErr: nil,
		},
		{
			give:       "update missing budget",
			giveID:     "task-999",
			giveStatus: StatusClosed,
			wantErr:    ErrBudgetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			budget := &TaskBudget{
				TaskID:      tt.giveID,
				TotalBudget: big.NewInt(1000000),
				Spent:       big.NewInt(0),
				Reserved:    big.NewInt(0),
				Status:      tt.giveStatus,
			}
			err := s.Update(budget)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			got, _ := s.Get(tt.giveID)
			assert.Equal(t, tt.giveStatus, got.Status)
		})
	}
}

func TestStore_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		giveID  string
		setup   func(*Store)
		wantErr error
	}{
		{
			give:   "delete existing budget",
			giveID: "task-1",
			setup: func(s *Store) {
				_, _ = s.Allocate("task-1", big.NewInt(1000000))
			},
			wantErr: nil,
		},
		{
			give:    "delete missing budget",
			giveID:  "task-999",
			wantErr: ErrBudgetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			err := s.Delete(tt.giveID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			_, err = s.Get(tt.giveID)
			assert.ErrorIs(t, err, ErrBudgetNotFound)
		})
	}
}

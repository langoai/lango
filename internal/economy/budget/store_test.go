package budget

import (
	"errors"
	"math/big"
	"testing"
)

func TestStore_Allocate(t *testing.T) {
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
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			got, err := s.Allocate(tt.giveID, big.NewInt(tt.giveTotal))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Allocate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Allocate() unexpected error: %v", err)
			}
			if got.TaskID != tt.giveID {
				t.Errorf("TaskID = %q, want %q", got.TaskID, tt.giveID)
			}
			if got.TotalBudget.Cmp(big.NewInt(tt.giveTotal)) != 0 {
				t.Errorf("TotalBudget = %s, want %d", got.TotalBudget, tt.giveTotal)
			}
			if got.Status != StatusActive {
				t.Errorf("Status = %q, want %q", got.Status, StatusActive)
			}
			if got.Spent.Sign() != 0 {
				t.Errorf("Spent = %s, want 0", got.Spent)
			}
			if got.Reserved.Sign() != 0 {
				t.Errorf("Reserved = %s, want 0", got.Reserved)
			}
		})
	}
}

func TestStore_Allocate_CopiesTotal(t *testing.T) {
	s := NewStore()
	total := big.NewInt(1000000)
	tb, err := s.Allocate("task-1", total)
	if err != nil {
		t.Fatalf("Allocate() unexpected error: %v", err)
	}

	total.SetInt64(0)
	if tb.TotalBudget.Cmp(big.NewInt(1000000)) != 0 {
		t.Errorf("TotalBudget mutated by caller: got %s", tb.TotalBudget)
	}
}

func TestStore_Get(t *testing.T) {
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
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			got, err := s.Get(tt.giveID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Get() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get() unexpected error: %v", err)
			}
			if got.TaskID != tt.giveID {
				t.Errorf("TaskID = %q, want %q", got.TaskID, tt.giveID)
			}
		})
	}
}

func TestStore_List(t *testing.T) {
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
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			got := s.List()
			if len(got) != tt.wantCount {
				t.Errorf("List() returned %d budgets, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestStore_Update(t *testing.T) {
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
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Update() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}

			got, _ := s.Get(tt.giveID)
			if got.Status != tt.giveStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.giveStatus)
			}
		})
	}
}

func TestStore_Delete(t *testing.T) {
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
			s := NewStore()
			if tt.setup != nil {
				tt.setup(s)
			}

			err := s.Delete(tt.giveID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Delete() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete() unexpected error: %v", err)
			}

			_, err = s.Get(tt.giveID)
			if !errors.Is(err, ErrBudgetNotFound) {
				t.Errorf("Get() after Delete() error = %v, want %v", err, ErrBudgetNotFound)
			}
		})
	}
}

package budget

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	ErrBudgetExists   = errors.New("budget already exists")
	ErrBudgetNotFound = errors.New("budget not found")
)

// Store is an in-memory store for task budgets.
type Store struct {
	mu      sync.RWMutex
	budgets map[string]*TaskBudget
}

// NewStore creates a new budget store.
func NewStore() *Store {
	return &Store{
		budgets: make(map[string]*TaskBudget),
	}
}

// Allocate creates a new task budget with the given total.
func (s *Store) Allocate(taskID string, total *big.Int) (*TaskBudget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.budgets[taskID]; exists {
		return nil, fmt.Errorf("allocate %q: %w", taskID, ErrBudgetExists)
	}

	now := time.Now()
	tb := &TaskBudget{
		TaskID:      taskID,
		TotalBudget: new(big.Int).Set(total),
		Spent:       new(big.Int),
		Reserved:    new(big.Int),
		Status:      StatusActive,
		Entries:     make([]SpendEntry, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.budgets[taskID] = tb

	return tb, nil
}

// Get returns the task budget for the given task ID.
func (s *Store) Get(taskID string) (*TaskBudget, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tb, exists := s.budgets[taskID]
	if !exists {
		return nil, fmt.Errorf("get %q: %w", taskID, ErrBudgetNotFound)
	}
	return tb, nil
}

// List returns all task budgets.
func (s *Store) List() []*TaskBudget {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*TaskBudget, 0, len(s.budgets))
	for _, tb := range s.budgets {
		result = append(result, tb)
	}
	return result
}

// Update replaces the stored budget with the provided one.
func (s *Store) Update(budget *TaskBudget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.budgets[budget.TaskID]; !exists {
		return fmt.Errorf("update %q: %w", budget.TaskID, ErrBudgetNotFound)
	}

	budget.UpdatedAt = time.Now()
	s.budgets[budget.TaskID] = budget
	return nil
}

// Delete removes the task budget for the given task ID.
func (s *Store) Delete(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.budgets[taskID]; !exists {
		return fmt.Errorf("delete %q: %w", taskID, ErrBudgetNotFound)
	}

	delete(s.budgets, taskID)
	return nil
}

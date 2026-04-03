package agentrt

import (
	"fmt"
	"sync"
	"time"
)

// AgentRunStore manages agent run lifecycle.
type AgentRunStore interface {
	Create(run *AgentRun) error
	Get(id string) (*AgentRun, error)
	List() []*AgentRun
	UpdateStatus(id string, status AgentRunStatus, result, errMsg string) error
	Cancel(id string) error
}

// InMemoryAgentRunStore is a thread-safe in-memory implementation of AgentRunStore.
type InMemoryAgentRunStore struct {
	mu   sync.RWMutex
	runs map[string]*AgentRun
}

// NewInMemoryAgentRunStore creates a new in-memory agent run store.
func NewInMemoryAgentRunStore() *InMemoryAgentRunStore {
	return &InMemoryAgentRunStore{
		runs: make(map[string]*AgentRun),
	}
}

// Compile-time interface check.
var _ AgentRunStore = (*InMemoryAgentRunStore)(nil)

// Create stores a new agent run. Returns an error if the ID already exists.
func (s *InMemoryAgentRunStore) Create(run *AgentRun) error {
	if run == nil {
		return fmt.Errorf("create agent run: nil run")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.runs[run.ID]; exists {
		return fmt.Errorf("create agent run: ID %q already exists", run.ID)
	}
	s.runs[run.ID] = run
	return nil
}

// Get returns a copy of the agent run with the given ID.
// Returns an error if the run is not found.
func (s *InMemoryAgentRunStore) Get(id string) (*AgentRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, ok := s.runs[id]
	if !ok {
		return nil, fmt.Errorf("get agent run: ID %q not found", id)
	}
	return copyRun(run), nil
}

// List returns copies of all agent runs.
func (s *InMemoryAgentRunStore) List() []*AgentRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AgentRun, 0, len(s.runs))
	for _, run := range s.runs {
		result = append(result, copyRun(run))
	}
	return result
}

// UpdateStatus updates the status, result, error, and completedAt fields of an agent run.
// Returns an error if the run is not found or is already in a terminal status.
func (s *InMemoryAgentRunStore) UpdateStatus(id string, status AgentRunStatus, result, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return fmt.Errorf("update agent run status: ID %q not found", id)
	}
	if run.Status.isTerminal() {
		return fmt.Errorf("update agent run status: ID %q is already %s", id, run.Status)
	}

	run.Status = status
	run.Result = result
	run.Error = errMsg
	if status.isTerminal() {
		run.CompletedAt = time.Now()
	}
	return nil
}

// Cancel cancels an agent run by calling its CancelFn (if set) and setting the
// status to Cancelled. Returns an error if the run is not found or is already
// in a terminal status. This follows the same guard pattern as background.Manager.
func (s *InMemoryAgentRunStore) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return fmt.Errorf("cancel agent run: ID %q not found", id)
	}
	if run.Status.isTerminal() {
		return fmt.Errorf("cancel agent run: ID %q is already %s", id, run.Status)
	}

	run.Status = AgentRunCancelled
	run.CompletedAt = time.Now()
	if run.CancelFn != nil {
		run.CancelFn()
	}
	return nil
}

// copyRun returns a shallow copy of the AgentRun with a copied AllowedTools
// slice. CancelFn is deliberately NOT copied to prevent external callers from
// invoking cancellation through a returned snapshot.
func copyRun(run *AgentRun) *AgentRun {
	cp := *run
	cp.CancelFn = nil
	if run.AllowedTools != nil {
		cp.AllowedTools = make([]string, len(run.AllowedTools))
		copy(cp.AllowedTools, run.AllowedTools)
	}
	return &cp
}

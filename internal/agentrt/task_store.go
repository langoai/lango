package agentrt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// TaskEntry represents a structured task for tracking work.
type TaskEntry struct {
	ID          string
	Title       string
	Status      string // "todo", "in_progress", "done", "blocked"
	AgentID     string // optional link to spawned agent
	ParentID    string // for hierarchical tasks
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaskStore manages task lifecycle.
type TaskStore interface {
	Create(entry *TaskEntry) error
	Get(id string) (*TaskEntry, error)
	List(statusFilter, parentFilter string) []*TaskEntry
	Update(id string, status, description string) error
}

// InMemoryTaskStore is a thread-safe in-memory implementation of TaskStore.
type InMemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*TaskEntry
}

// NewInMemoryTaskStore creates a new in-memory task store.
func NewInMemoryTaskStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks: make(map[string]*TaskEntry),
	}
}

// Compile-time interface check.
var _ TaskStore = (*InMemoryTaskStore)(nil)

// Create stores a new task entry. Returns an error if the entry is nil or the ID already exists.
func (s *InMemoryTaskStore) Create(entry *TaskEntry) error {
	if entry == nil {
		return fmt.Errorf("create task: nil entry")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[entry.ID]; exists {
		return fmt.Errorf("create task: ID %q already exists", entry.ID)
	}
	s.tasks[entry.ID] = entry
	return nil
}

// Get returns a copy of the task entry with the given ID.
// Returns an error if the task is not found.
func (s *InMemoryTaskStore) Get(id string) (*TaskEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("get task: ID %q not found", id)
	}
	return copyTask(task), nil
}

// List returns copies of tasks matching the optional filters.
// Empty filter strings match all tasks.
func (s *InMemoryTaskStore) List(statusFilter, parentFilter string) []*TaskEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*TaskEntry, 0, len(s.tasks))
	for _, task := range s.tasks {
		if statusFilter != "" && task.Status != statusFilter {
			continue
		}
		if parentFilter != "" && task.ParentID != parentFilter {
			continue
		}
		result = append(result, copyTask(task))
	}
	return result
}

// Update updates the status and/or description of a task.
// Returns an error if the task is not found.
// Empty strings are treated as "no change" for each field.
func (s *InMemoryTaskStore) Update(id string, status, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("update task: ID %q not found", id)
	}
	if status != "" {
		task.Status = status
	}
	if description != "" {
		task.Description = description
	}
	task.UpdatedAt = time.Now()
	return nil
}

// generateTaskID creates a random hex ID for a task.
func generateTaskID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate task ID: %w", err)
	}
	return "task-" + hex.EncodeToString(b), nil
}

// copyTask returns a shallow copy of the TaskEntry.
func copyTask(task *TaskEntry) *TaskEntry {
	cp := *task
	return &cp
}

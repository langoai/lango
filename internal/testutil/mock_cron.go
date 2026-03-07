package testutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/langoai/lango/internal/cron"
)

// Compile-time interface check.
var _ cron.Store = (*MockCronStore)(nil)

// MockCronStore is a thread-safe in-memory mock of cron.Store.
type MockCronStore struct {
	mu      sync.Mutex
	jobs    map[string]cron.Job
	history []cron.HistoryEntry

	CreateErr      error
	GetErr         error
	ListErr        error
	UpdateErr      error
	DeleteErr      error
	SaveHistoryErr error

	createCalls int
}

// NewMockCronStore creates an empty MockCronStore.
func NewMockCronStore() *MockCronStore {
	return &MockCronStore{
		jobs: make(map[string]cron.Job),
	}
}

func (m *MockCronStore) Create(_ context.Context, job cron.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *MockCronStore) Get(_ context.Context, id string) (*cron.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	job, ok := m.jobs[id]
	if !ok {
		return nil, fmt.Errorf("job %q not found", id)
	}
	return &job, nil
}

func (m *MockCronStore) GetByName(_ context.Context, name string) (*cron.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	for _, job := range m.jobs {
		if job.Name == name {
			return &job, nil
		}
	}
	return nil, fmt.Errorf("job %q not found", name)
}

func (m *MockCronStore) List(_ context.Context) ([]cron.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	result := make([]cron.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		result = append(result, job)
	}
	return result, nil
}

func (m *MockCronStore) ListEnabled(_ context.Context) ([]cron.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	var result []cron.Job
	for _, job := range m.jobs {
		if job.Enabled {
			result = append(result, job)
		}
	}
	return result, nil
}

func (m *MockCronStore) Update(_ context.Context, job cron.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *MockCronStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.jobs, id)
	return nil
}

func (m *MockCronStore) SaveHistory(_ context.Context, entry cron.HistoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.SaveHistoryErr != nil {
		return m.SaveHistoryErr
	}
	m.history = append(m.history, entry)
	return nil
}

func (m *MockCronStore) ListHistory(_ context.Context, jobID string, limit int) ([]cron.HistoryEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []cron.HistoryEntry
	for _, h := range m.history {
		if h.JobID == jobID {
			result = append(result, h)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockCronStore) ListAllHistory(_ context.Context, limit int) ([]cron.HistoryEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]cron.HistoryEntry, len(m.history))
	copy(result, m.history)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// CreateCalls returns the number of Create calls.
func (m *MockCronStore) CreateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createCalls
}

// JobCount returns the number of stored jobs.
func (m *MockCronStore) JobCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.jobs)
}

// HistoryCount returns the number of stored history entries.
func (m *MockCronStore) HistoryCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.history)
}

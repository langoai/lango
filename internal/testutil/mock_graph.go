package testutil

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/graph"
)

// Compile-time interface check.
var _ graph.Store = (*MockGraphStore)(nil)

// MockGraphStore is a thread-safe in-memory mock of graph.Store.
type MockGraphStore struct {
	mu      sync.Mutex
	triples []graph.Triple

	AddErr   error
	QueryErr error
	addCalls int
}

// NewMockGraphStore creates an empty MockGraphStore.
func NewMockGraphStore() *MockGraphStore {
	return &MockGraphStore{}
}

func (m *MockGraphStore) AddTriple(_ context.Context, t graph.Triple) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addCalls++
	if m.AddErr != nil {
		return m.AddErr
	}
	m.triples = append(m.triples, t)
	return nil
}

func (m *MockGraphStore) AddTriples(_ context.Context, triples []graph.Triple) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addCalls++
	if m.AddErr != nil {
		return m.AddErr
	}
	m.triples = append(m.triples, triples...)
	return nil
}

func (m *MockGraphStore) RemoveTriple(_ context.Context, t graph.Triple) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, tr := range m.triples {
		if tr.Subject == t.Subject && tr.Predicate == t.Predicate && tr.Object == t.Object {
			m.triples = append(m.triples[:i], m.triples[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockGraphStore) QueryBySubject(_ context.Context, subject string) ([]graph.Triple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.QueryErr != nil {
		return nil, m.QueryErr
	}
	var result []graph.Triple
	for _, t := range m.triples {
		if t.Subject == subject {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockGraphStore) QueryByObject(_ context.Context, object string) ([]graph.Triple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.QueryErr != nil {
		return nil, m.QueryErr
	}
	var result []graph.Triple
	for _, t := range m.triples {
		if t.Object == object {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockGraphStore) QueryBySubjectPredicate(_ context.Context, subject, predicate string) ([]graph.Triple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.QueryErr != nil {
		return nil, m.QueryErr
	}
	var result []graph.Triple
	for _, t := range m.triples {
		if t.Subject == subject && t.Predicate == predicate {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockGraphStore) Traverse(_ context.Context, startNode string, maxDepth int, predicates []string) ([]graph.Triple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.QueryErr != nil {
		return nil, m.QueryErr
	}
	_ = maxDepth
	_ = predicates
	var result []graph.Triple
	for _, t := range m.triples {
		if t.Subject == startNode || t.Object == startNode {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockGraphStore) Count(_ context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.triples), nil
}

func (m *MockGraphStore) PredicateStats(_ context.Context) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := make(map[string]int)
	for _, t := range m.triples {
		stats[t.Predicate]++
	}
	return stats, nil
}

func (m *MockGraphStore) AllTriples(_ context.Context) ([]graph.Triple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]graph.Triple, len(m.triples))
	copy(result, m.triples)
	return result, nil
}

func (m *MockGraphStore) ClearAll(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.triples = nil
	return nil
}

func (m *MockGraphStore) Close() error { return nil }

// AddCalls returns the number of Add calls.
func (m *MockGraphStore) AddCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.addCalls
}

// TripleCount returns the number of stored triples.
func (m *MockGraphStore) TripleCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.triples)
}

package testutil

import (
	"context"
	"iter"
	"sync"

	"github.com/langoai/lango/internal/provider"
)

// Compile-time interface check.
var _ provider.Provider = (*MockProvider)(nil)

// MockProvider is a thread-safe mock implementation of provider.Provider for tests.
type MockProvider struct {
	mu sync.Mutex

	// Configurable responses
	ProviderID string
	Events     []provider.StreamEvent
	Models     []provider.ModelInfo

	// Configurable error injection
	GenerateErr   error
	ListModelsErr error

	// Call tracking
	generateCalls   int
	listModelsCalls int
	lastParams      *provider.GenerateParams
}

// NewMockProvider creates a new MockProvider with the given ID.
func NewMockProvider(id string) *MockProvider {
	return &MockProvider{
		ProviderID: id,
		Events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "mock response"},
			{Type: provider.StreamEventDone},
		},
	}
}

func (m *MockProvider) ID() string {
	return m.ProviderID
}

func (m *MockProvider) Generate(_ context.Context, params provider.GenerateParams) (iter.Seq2[provider.StreamEvent, error], error) {
	m.mu.Lock()
	m.generateCalls++
	cp := params
	m.lastParams = &cp
	events := make([]provider.StreamEvent, len(m.Events))
	copy(events, m.Events)
	genErr := m.GenerateErr
	m.mu.Unlock()

	if genErr != nil {
		return nil, genErr
	}

	return func(yield func(provider.StreamEvent, error) bool) {
		for _, ev := range events {
			if !yield(ev, nil) {
				return
			}
		}
	}, nil
}

func (m *MockProvider) ListModels(_ context.Context) ([]provider.ModelInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listModelsCalls++
	if m.ListModelsErr != nil {
		return nil, m.ListModelsErr
	}
	result := make([]provider.ModelInfo, len(m.Models))
	copy(result, m.Models)
	return result, nil
}

// Inspection methods

// GenerateCalls returns the number of Generate calls.
func (m *MockProvider) GenerateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.generateCalls
}

// ListModelsCalls returns the number of ListModels calls.
func (m *MockProvider) ListModelsCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listModelsCalls
}

// LastParams returns the last GenerateParams passed to Generate.
func (m *MockProvider) LastParams() *provider.GenerateParams {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastParams == nil {
		return nil
	}
	cp := *m.lastParams
	return &cp
}

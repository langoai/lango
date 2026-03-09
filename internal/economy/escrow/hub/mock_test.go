package hub

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/contract"
)

// mockCaller implements contract.ContractCaller for unit tests.
type mockCaller struct {
	mu sync.Mutex

	readResult  *contract.ContractCallResult
	readErr     error
	writeResult *contract.ContractCallResult
	writeErr    error

	readCalls  []contract.ContractCallRequest
	writeCalls []contract.ContractCallRequest
}

func newMockCaller() *mockCaller {
	return &mockCaller{
		readResult: &contract.ContractCallResult{},
		writeResult: &contract.ContractCallResult{
			TxHash: "0xmocktx",
		},
	}
}

func (m *mockCaller) Read(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readCalls = append(m.readCalls, req)
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.readResult, nil
}

func (m *mockCaller) Write(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeCalls = append(m.writeCalls, req)
	if m.writeErr != nil {
		return nil, m.writeErr
	}
	return m.writeResult, nil
}

// mockOnChainStore implements OnChainStore for tests.
type mockOnChainStore struct {
	mu      sync.RWMutex
	mapping map[string]string // dealID → escrowID
}

func newMockOnChainStore() *mockOnChainStore {
	return &mockOnChainStore{
		mapping: make(map[string]string),
	}
}

func (s *mockOnChainStore) GetByOnChainDealID(dealID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.mapping[dealID]
	if !ok {
		return "", nil
	}
	return id, nil
}

func (s *mockOnChainStore) Set(dealID, escrowID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mapping[dealID] = escrowID
}

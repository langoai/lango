package testutil

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/security"
)

// Compile-time interface check.
var _ security.CryptoProvider = (*MockCryptoProvider)(nil)

// MockCryptoProvider is a thread-safe mock of security.CryptoProvider.
type MockCryptoProvider struct {
	mu sync.Mutex

	SignResult    []byte
	EncryptResult []byte
	DecryptResult []byte

	SignErr    error
	EncryptErr error
	DecryptErr error

	signCalls    int
	encryptCalls int
	decryptCalls int
}

// NewMockCryptoProvider creates a MockCryptoProvider with default passthrough behavior.
func NewMockCryptoProvider() *MockCryptoProvider {
	return &MockCryptoProvider{
		SignResult:    []byte("mock-signature"),
		EncryptResult: []byte("mock-ciphertext"),
		DecryptResult: []byte("mock-plaintext"),
	}
}

func (m *MockCryptoProvider) Sign(_ context.Context, _ string, _ []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signCalls++
	if m.SignErr != nil {
		return nil, m.SignErr
	}
	return m.SignResult, nil
}

func (m *MockCryptoProvider) Encrypt(_ context.Context, _ string, _ []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptCalls++
	if m.EncryptErr != nil {
		return nil, m.EncryptErr
	}
	return m.EncryptResult, nil
}

func (m *MockCryptoProvider) Decrypt(_ context.Context, _ string, _ []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decryptCalls++
	if m.DecryptErr != nil {
		return nil, m.DecryptErr
	}
	return m.DecryptResult, nil
}

// SignCalls returns the number of Sign calls.
func (m *MockCryptoProvider) SignCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signCalls
}

// EncryptCalls returns the number of Encrypt calls.
func (m *MockCryptoProvider) EncryptCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.encryptCalls
}

// DecryptCalls returns the number of Decrypt calls.
func (m *MockCryptoProvider) DecryptCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.decryptCalls
}

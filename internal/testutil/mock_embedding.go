package testutil

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/embedding"
)

// Compile-time interface check.
var _ embedding.EmbeddingProvider = (*MockEmbeddingProvider)(nil)

// MockEmbeddingProvider is a thread-safe mock of embedding.EmbeddingProvider.
type MockEmbeddingProvider struct {
	mu sync.Mutex

	ProviderID     string
	EmbedDimension int
	Vectors        [][]float32

	EmbedErr   error
	embedCalls int
	lastTexts  []string
}

// NewMockEmbeddingProvider creates a provider that returns zero vectors of the given dimension.
func NewMockEmbeddingProvider(id string, dims int) *MockEmbeddingProvider {
	return &MockEmbeddingProvider{
		ProviderID:     id,
		EmbedDimension: dims,
	}
}

func (m *MockEmbeddingProvider) ID() string { return m.ProviderID }

func (m *MockEmbeddingProvider) Embed(_ context.Context, texts []string) ([][]float32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embedCalls++
	m.lastTexts = texts
	if m.EmbedErr != nil {
		return nil, m.EmbedErr
	}
	if m.Vectors != nil {
		return m.Vectors, nil
	}
	result := make([][]float32, len(texts))
	for i := range result {
		result[i] = make([]float32, m.EmbedDimension)
	}
	return result, nil
}

func (m *MockEmbeddingProvider) Dimensions() int { return m.EmbedDimension }

// EmbedCalls returns the number of Embed calls.
func (m *MockEmbeddingProvider) EmbedCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.embedCalls
}

// LastTexts returns the last texts passed to Embed.
func (m *MockEmbeddingProvider) LastTexts() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.lastTexts))
	copy(cp, m.lastTexts)
	return cp
}

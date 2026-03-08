package embedding

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockProvider is a test EmbeddingProvider that returns simple embeddings.
type mockProvider struct {
	dim        int
	embedCalls int
}

func (m *mockProvider) ID() string      { return "mock" }
func (m *mockProvider) Dimensions() int { return m.dim }

func (m *mockProvider) Embed(_ context.Context, texts []string) ([][]float32, error) {
	m.embedCalls++
	result := make([][]float32, len(texts))
	for i := range texts {
		vec := make([]float32, m.dim)
		// Simple deterministic embedding: set index i%dim to 1.
		vec[i%m.dim] = 1.0
		result[i] = vec
	}
	return result, nil
}

// mockStore records upserted records.
type mockStore struct {
	mu      sync.Mutex
	records []VectorRecord
}

func (s *mockStore) Upsert(_ context.Context, recs []VectorRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, recs...)
	return nil
}

func (s *mockStore) Search(_ context.Context, _ string, _ []float32, _ int, _ *SearchOptions) ([]SearchResult, error) {
	return nil, nil
}

func (s *mockStore) Delete(_ context.Context, _ string, _ []string) error {
	return nil
}

func (s *mockStore) Close() error { return nil }

func (s *mockStore) getRecords() []VectorRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]VectorRecord, len(s.records))
	copy(cp, s.records)
	return cp
}

func TestEmbeddingBuffer_ProcessesRequests(t *testing.T) {
	provider := &mockProvider{dim: 4}
	store := &mockStore{}
	logger := zap.NewNop().Sugar()

	buf := NewEmbeddingBuffer(provider, store, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	buf.Enqueue(EmbedRequest{ID: "k1", Collection: "knowledge", Content: "hello world"})
	buf.Enqueue(EmbedRequest{ID: "k2", Collection: "knowledge", Content: "goodbye world"})

	// Give the buffer time to flush.
	time.Sleep(3 * time.Second)
	buf.Stop()
	wg.Wait()

	records := store.getRecords()
	require.Len(t, records, 2)
	assert.Equal(t, "k1", records[0].ID)
	assert.Equal(t, "k2", records[1].ID)
	assert.Equal(t, 4, len(records[0].Embedding))
}

func TestEmbeddingBuffer_GracefulShutdown(t *testing.T) {
	provider := &mockProvider{dim: 4}
	store := &mockStore{}
	logger := zap.NewNop().Sugar()

	buf := NewEmbeddingBuffer(provider, store, logger)

	var wg sync.WaitGroup
	buf.Start(&wg)

	// Enqueue and immediately stop — should drain.
	for i := 0; i < 10; i++ {
		buf.Enqueue(EmbedRequest{ID: "item", Collection: "knowledge", Content: "content"})
	}

	buf.Stop()
	wg.Wait()

	records := store.getRecords()
	assert.Len(t, records, 10)
}

//go:build vec

package embedding

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// staticResolver returns the source_id as content.
type staticResolver struct{}

func (r *staticResolver) ResolveContent(_ context.Context, collection, id string) (string, error) {
	return fmt.Sprintf("content for %s/%s", collection, id), nil
}

func setupRAGTest(t *testing.T) (*RAGService, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	store, err := NewSQLiteVecStore(db, 4)
	require.NoError(t, err)

	provider := &mockProvider{dim: 4}
	resolver := &staticResolver{}
	logger := zap.NewNop().Sugar()

	svc := NewRAGService(provider, store, resolver, logger)

	// Seed data.
	ctx := context.Background()
	require.NoError(t, store.Upsert(ctx, []VectorRecord{
		{ID: "k1", Collection: "knowledge", Embedding: []float32{1, 0, 0, 0}},
		{ID: "k2", Collection: "knowledge", Embedding: []float32{0, 1, 0, 0}},
		{ID: "o1", Collection: "observation", Embedding: []float32{0, 0, 1, 0}},
	}))

	t.Cleanup(func() { db.Close() })
	return svc, db
}

func TestRAGService_Retrieve(t *testing.T) {
	svc, _ := setupRAGTest(t)
	ctx := context.Background()

	results, err := svc.Retrieve(ctx, "test query", RetrieveOptions{
		Limit: 3,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// All results should have resolved content.
	for _, r := range results {
		assert.Contains(t, r.Content, "content for")
	}
}

func TestRAGService_RetrieveFilteredCollection(t *testing.T) {
	svc, _ := setupRAGTest(t)
	ctx := context.Background()

	results, err := svc.Retrieve(ctx, "test query", RetrieveOptions{
		Collections: []string{"observation"},
		Limit:       5,
	})
	require.NoError(t, err)
	for _, r := range results {
		assert.Equal(t, "observation", r.Collection)
	}
}

func TestRAGService_RetrieveEmptyQuery(t *testing.T) {
	svc, _ := setupRAGTest(t)
	ctx := context.Background()

	results, err := svc.Retrieve(ctx, "", RetrieveOptions{Limit: 5})
	require.NoError(t, err)
	assert.Empty(t, results)
}

//go:build vec

package embedding

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteVecStore_UpsertAndSearch(t *testing.T) {
	db := openTestDB(t)
	store, err := NewSQLiteVecStore(db, 4)
	require.NoError(t, err)

	ctx := context.Background()

	// Insert records.
	records := []VectorRecord{
		{ID: "k1", Collection: "knowledge", Embedding: []float32{1, 0, 0, 0}, Metadata: map[string]string{"category": "fact"}},
		{ID: "k2", Collection: "knowledge", Embedding: []float32{0, 1, 0, 0}, Metadata: map[string]string{"category": "rule"}},
		{ID: "o1", Collection: "observation", Embedding: []float32{0, 0, 1, 0}, Metadata: map[string]string{"session_key": "s1"}},
	}
	require.NoError(t, store.Upsert(ctx, records))

	// Search knowledge collection — closest to [1,0,0,0] should be k1.
	results, err := store.Search(ctx, "knowledge", []float32{0.9, 0.1, 0, 0}, 2, nil)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "k1", results[0].ID)
	assert.Equal(t, "knowledge", results[0].Collection)
	assert.Equal(t, "fact", results[0].Metadata["category"])

	// Search observation collection.
	results, err = store.Search(ctx, "observation", []float32{0, 0, 0.8, 0.2}, 1, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "o1", results[0].ID)
}

func TestSQLiteVecStore_Upsert_Replaces(t *testing.T) {
	db := openTestDB(t)
	store, err := NewSQLiteVecStore(db, 4)
	require.NoError(t, err)

	ctx := context.Background()

	// Insert then upsert with a different embedding.
	require.NoError(t, store.Upsert(ctx, []VectorRecord{
		{ID: "k1", Collection: "knowledge", Embedding: []float32{1, 0, 0, 0}},
	}))

	require.NoError(t, store.Upsert(ctx, []VectorRecord{
		{ID: "k1", Collection: "knowledge", Embedding: []float32{0, 1, 0, 0}},
	}))

	// Searching near [0,1,0,0] should still find k1 with updated embedding.
	results, err := store.Search(ctx, "knowledge", []float32{0, 1, 0, 0}, 5, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "k1", results[0].ID)
}

func TestSQLiteVecStore_Delete(t *testing.T) {
	db := openTestDB(t)
	store, err := NewSQLiteVecStore(db, 4)
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, store.Upsert(ctx, []VectorRecord{
		{ID: "k1", Collection: "knowledge", Embedding: []float32{1, 0, 0, 0}},
		{ID: "k2", Collection: "knowledge", Embedding: []float32{0, 1, 0, 0}},
	}))

	require.NoError(t, store.Delete(ctx, "knowledge", []string{"k1"}))

	results, err := store.Search(ctx, "knowledge", []float32{1, 0, 0, 0}, 5, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "k2", results[0].ID)
}

func TestSQLiteVecStore_EmptyOperations(t *testing.T) {
	db := openTestDB(t)
	store, err := NewSQLiteVecStore(db, 4)
	require.NoError(t, err)

	ctx := context.Background()

	// Empty upsert should be no-op.
	require.NoError(t, store.Upsert(ctx, nil))

	// Empty delete should be no-op.
	require.NoError(t, store.Delete(ctx, "knowledge", nil))

	// Search on empty store should return empty.
	results, err := store.Search(ctx, "knowledge", []float32{1, 0, 0, 0}, 5, nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

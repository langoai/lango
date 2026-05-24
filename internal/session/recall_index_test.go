package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/types"
)

func skipWithoutFTS5Recall(t *testing.T, store *EntStore) {
	t.Helper()
	if !search.ProbeFTS5(store.DB()) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}
}

func TestRecallIndex_IndexAndSearch(t *testing.T) {
	store := newTestEntStore(t)
	skipWithoutFTS5Recall(t, store)
	require.NoError(t, store.Create(&Session{
		Key: "sess-1",
		History: []Message{
			{Role: types.RoleUser, Content: "we decided to use PostgreSQL for the analytics pipeline"},
			{Role: types.RoleAssistant, Content: "acknowledged; keeping the Postgres decision for downstream tasks"},
		},
	}))

	idx := NewRecallIndex(store)
	require.NoError(t, idx.IndexSession(context.Background(), "sess-1"))

	results, err := idx.Search(context.Background(), "postgresql analytics", 5)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, "sess-1", results[0].RowID)
}

func TestRecallIndex_UpsertReplacesRow(t *testing.T) {
	store := newTestEntStore(t)
	skipWithoutFTS5Recall(t, store)
	require.NoError(t, store.Create(&Session{
		Key: "sess-1",
		History: []Message{
			{Role: types.RoleUser, Content: "first conversation about MongoDB"},
		},
	}))

	idx := NewRecallIndex(store)
	require.NoError(t, idx.IndexSession(context.Background(), "sess-1"))

	// Re-index after new messages: upsert should replace the row.
	require.NoError(t, store.AppendMessage("sess-1", Message{
		Role: types.RoleAssistant, Content: "switched to Redis for caching layer",
	}))
	require.NoError(t, idx.IndexSession(context.Background(), "sess-1"))

	// Count rows to confirm single-row-per-session invariant.
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM " + RecallTableName + " WHERE source_id = 'sess-1'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRecallIndex_ProcessPending(t *testing.T) {
	store := newTestEntStore(t)
	skipWithoutFTS5Recall(t, store)
	require.NoError(t, store.Create(&Session{
		Key: "sess-a",
		History: []Message{
			{Role: types.RoleUser, Content: "talking about Kubernetes"},
		},
	}))
	require.NoError(t, store.Create(&Session{Key: "sess-b"}))

	require.NoError(t, store.MarkEndPending("sess-a"))
	require.NoError(t, store.MarkEndPending("sess-b"))

	idx := NewRecallIndex(store)
	require.NoError(t, idx.ProcessPending(context.Background()))

	pending, err := store.ListEndPending()
	require.NoError(t, err)
	assert.Empty(t, pending, "flag cleared after successful processing")

	// sess-a indexed, sess-b had no messages so no row, but flag still cleared.
	results, err := idx.Search(context.Background(), "kubernetes", 5)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "sess-a", results[0].RowID)
}

func TestRecallIndex_RedactsSensitiveProjection(t *testing.T) {
	store := newTestEntStore(t)
	skipWithoutFTS5Recall(t, store)
	require.NoError(t, store.Create(&Session{
		Key: "sess-secret",
		History: []Message{
			{Role: types.RoleUser, Content: "email alice@example.com token SECRETSECRETSECRETSECRETSECRETSECRET 123456789"},
		},
	}))

	idx := NewRecallIndex(store)
	require.NoError(t, idx.IndexSession(context.Background(), "sess-secret"))

	summary, err := idx.GetSummary(context.Background(), "sess-secret")
	require.NoError(t, err)
	assert.NotContains(t, summary, "alice@example.com")
	assert.NotContains(t, summary, "SECRETSECRETSECRETSECRETSECRETSECRET")
	assert.NotContains(t, summary, "123456789")
	assert.Contains(t, summary, "[email]")
	assert.Contains(t, summary, "[secret]")
	assert.Contains(t, summary, "[number]")
}

func TestRecallIndex_ProtectedDecryptFailureDoesNotUseProjection(t *testing.T) {
	store := newTestEntStore(t, WithPayloadProtector(stubPayloadProtector{}))
	skipWithoutFTS5Recall(t, store)
	require.NoError(t, store.Create(&Session{
		Key: "sess-protected-recall",
		History: []Message{
			{Role: types.RoleUser, Content: "top secret"},
		},
	}))
	store.SetPayloadProtector(failDecryptProtector{})

	idx := NewRecallIndex(store)
	err := idx.IndexSession(context.Background(), "sess-protected-recall")
	require.Error(t, err)
}

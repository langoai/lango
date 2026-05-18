package knowledge

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/search"
)

func TestResolveKnowledgeByKeys_PreservesFTSOrderAndFiltersCategory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entries := []KnowledgeEntry{
		{
			Key:      "knowledge-fact-a",
			Category: entknowledge.CategoryFact,
			Content:  "first fact",
		},
		{
			Key:      "knowledge-rule",
			Category: entknowledge.CategoryRule,
			Content:  "filtered rule",
		},
		{
			Key:      "knowledge-fact-b",
			Category: entknowledge.CategoryFact,
			Content:  "second fact",
		},
	}
	for _, entry := range entries {
		require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))
	}

	got, err := store.resolveKnowledgeByKeys(ctx, []search.SearchResult{
		{RowID: "missing"},
		{RowID: "knowledge-rule"},
		{RowID: "knowledge-fact-b"},
		{RowID: "knowledge-fact-a"},
	}, string(entknowledge.CategoryFact))
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "knowledge-fact-b", got[0].Key)
	require.Equal(t, "knowledge-fact-a", got[1].Key)
	require.Equal(t, entknowledge.CategoryFact, got[0].Category)
	require.Equal(t, entknowledge.CategoryFact, got[1].Category)
}

func TestResolveKnowledgeScoredByKeys_PreservesFTSOrderScoresAndFiltersCategory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "scored-fact-a",
		Category: entknowledge.CategoryFact,
		Content:  "alpha fact",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "scored-rule",
		Category: entknowledge.CategoryRule,
		Content:  "filtered rule",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "scored-fact-b",
		Category: entknowledge.CategoryFact,
		Content:  "beta fact",
	}))

	got, err := store.resolveKnowledgeScoredByKeys(ctx, []search.SearchResult{
		{RowID: "scored-rule", Rank: -0.01},
		{RowID: "scored-fact-b", Rank: -0.25},
		{RowID: "scored-fact-a", Rank: -0.75},
	}, string(entknowledge.CategoryFact))
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "scored-fact-b", got[0].Entry.Key)
	require.Equal(t, 0.25, got[0].Score)
	require.Equal(t, "fts5", got[0].SearchSource)
	require.Equal(t, "scored-fact-a", got[1].Entry.Key)
	require.Equal(t, 0.75, got[1].Score)
	require.Equal(t, "fts5", got[1].SearchSource)
}

func TestSearchKnowledge_FTS5EmptyResultDoesNotFallbackUntilIndexCleared(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "fts-empty-knowledge",
		Category: entknowledge.CategoryFact,
		Content:  "needle only present in ent fallback",
	}))
	_, err := rawDB.ExecContext(ctx, `DELETE FROM knowledge_fts WHERE source_id = ?`, "fts-empty-knowledge")
	require.NoError(t, err)

	got, err := store.SearchKnowledge(ctx, "needle", "", 10)
	require.NoError(t, err)
	require.Empty(t, got)

	scored, err := store.SearchKnowledgeScored(ctx, "needle", "", 10)
	require.NoError(t, err)
	require.Empty(t, scored)

	store.SetFTS5Index(nil)
	got, err = store.SearchKnowledge(ctx, "needle", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "fts-empty-knowledge", got[0].Key)

	scored, err = store.SearchKnowledgeScored(ctx, "needle", "", 10)
	require.NoError(t, err)
	require.Len(t, scored, 1)
	require.Equal(t, "fts-empty-knowledge", scored[0].Entry.Key)
	require.Equal(t, "like", scored[0].SearchSource)
}

func TestResolveLearningsByIDs_PreservesFTSOrderAndFiltersCategory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "timeout one",
		ErrorPattern: "context deadline exceeded",
		Fix:          "increase timeout",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "permission one",
		ErrorPattern: "permission denied",
		Fix:          "change permissions",
		Category:     entlearning.CategoryPermission,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "timeout two",
		ErrorPattern: "client timeout",
		Fix:          "retry later",
		Category:     entlearning.CategoryTimeout,
	}))

	ids := learningIDsByTrigger(t, store, ctx)
	got, err := store.resolveLearningsByIDs(ctx, []search.SearchResult{
		{RowID: "not-a-uuid"},
		{RowID: ids["permission one"].String()},
		{RowID: ids["timeout two"].String()},
		{RowID: ids["timeout one"].String()},
	}, string(entlearning.CategoryTimeout))
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "timeout two", got[0].Trigger)
	require.Equal(t, "timeout one", got[1].Trigger)
	require.Equal(t, entlearning.CategoryTimeout, got[0].Category)
	require.Equal(t, entlearning.CategoryTimeout, got[1].Category)
}

func TestSearchLearnings_FTS5EmptyResultDoesNotFallbackUntilIndexCleared(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "learning fts empty",
		ErrorPattern: "panic stack trace",
		Fix:          "inspect nil pointer",
		Category:     entlearning.CategoryToolError,
	}))
	ids := learningIDsByTrigger(t, store, ctx)
	id := ids["learning fts empty"]
	_, err := rawDB.ExecContext(ctx, `DELETE FROM learning_fts WHERE source_id = ?`, id.String())
	require.NoError(t, err)

	got, err := store.SearchLearnings(ctx, "panic", "", 10)
	require.NoError(t, err)
	require.Empty(t, got)

	store.SetLearningFTS5Index(nil)
	got, err = store.SearchLearnings(ctx, "panic", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "learning fts empty", got[0].Trigger)
}

func TestSearchLearnings_LIKEFallbackFiltersAndOrdersByConfidence(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "low timeout",
		ErrorPattern: "timeout waiting for service",
		Fix:          "restart service",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "high timeout",
		ErrorPattern: "timeout waiting for database",
		Fix:          "increase pool",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "permission denied",
		ErrorPattern: "timeout word in unrelated category",
		Fix:          "fix permissions",
		Category:     entlearning.CategoryPermission,
	}))

	ids := learningIDsByTrigger(t, store, ctx)
	_, err := store.client.Learning.UpdateOneID(ids["high timeout"]).SetConfidence(0.9).Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.UpdateOneID(ids["low timeout"]).SetConfidence(0.2).Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchLearnings(ctx, "timeout", string(entlearning.CategoryTimeout), 10)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "high timeout", got[0].Trigger)
	require.Equal(t, "low timeout", got[1].Trigger)
}

func TestDeleteLearning_RemovesEntityAndFTS5Row(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "delete learning",
		ErrorPattern: "delete me",
		Fix:          "remove stale learning",
		Category:     entlearning.CategoryToolError,
	}))
	id := learningIDsByTrigger(t, store, ctx)["delete learning"]

	require.NoError(t, store.DeleteLearning(ctx, id))

	_, err := store.GetLearning(ctx, id)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLearningNotFound))

	var count int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM learning_fts WHERE source_id = ?`,
		id.String(),
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	err = store.DeleteLearning(ctx, uuid.New())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLearningNotFound))
}

func learningIDsByTrigger(
	t *testing.T,
	store *Store,
	ctx context.Context,
) map[string]uuid.UUID {
	t.Helper()

	rows, err := store.client.Learning.Query().All(ctx)
	require.NoError(t, err)

	ids := make(map[string]uuid.UUID, len(rows))
	for _, row := range rows {
		ids[row.Trigger] = row.ID
	}
	return ids
}

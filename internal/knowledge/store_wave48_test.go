package knowledge

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
)

func TestWave48KnowledgeEmptyQueryWithFTSUsesLIKEFallbackAndLatestFilters(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	for i := 0; i < 12; i++ {
		require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
			Key:      fmt.Sprintf("wave48-fact-%02d", i),
			Category: entknowledge.CategoryFact,
			Content:  fmt.Sprintf("wave48 fact content %02d", i),
		}))
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-versioned-category",
		Category: entknowledge.CategoryFact,
		Content:  "old fact content should not be latest",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-versioned-category",
		Category: entknowledge.CategoryRule,
		Content:  "current rule content",
	}))

	got, err := store.SearchKnowledge(ctx, "", string(entknowledge.CategoryFact), 0)
	require.NoError(t, err)
	require.Len(t, got, 10)
	for _, entry := range got {
		require.Equal(t, entknowledge.CategoryFact, entry.Category)
		require.NotEqual(t, "wave48-versioned-category", entry.Key)
	}

	_, err = store.GetKnowledge(ctx, "wave48-missing")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))
}

func TestWave48KnowledgeScoredEmptyQueryWithFTSUsesLIKEScores(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-score-low",
		Category: entknowledge.CategoryFact,
		Content:  "scored empty query low",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-score-high",
		Category: entknowledge.CategoryFact,
		Content:  "scored empty query high",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-score-filtered",
		Category: entknowledge.CategoryRule,
		Content:  "scored empty query filtered",
	}))
	_, err := store.client.Knowledge.Update().
		Where(entknowledge.Key("wave48-score-high"), entknowledge.IsLatest(true)).
		SetRelevanceScore(7.5).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Knowledge.Update().
		Where(entknowledge.Key("wave48-score-low"), entknowledge.IsLatest(true)).
		SetRelevanceScore(2.25).
		Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchKnowledgeScored(ctx, "", string(entknowledge.CategoryFact), 2)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "wave48-score-high", got[0].Entry.Key)
	require.Equal(t, 7.5, got[0].Score)
	require.Equal(t, "like", got[0].SearchSource)
	require.Equal(t, "wave48-score-low", got[1].Entry.Key)
	require.Equal(t, 2.25, got[1].Score)
	require.Equal(t, "like", got[1].SearchSource)
}

func TestWave48KnowledgeFTSLimitAndDeleteSync(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-delete-fts",
		Category: entknowledge.CategoryFact,
		Content:  "wave48limit token first version",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-delete-fts",
		Category: entknowledge.CategoryFact,
		Content:  "wave48limit token current version",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave48-keep-fts",
		Category: entknowledge.CategoryFact,
		Content:  "wave48limit token sibling",
	}))

	limited, err := store.SearchKnowledge(ctx, "wave48limit", "", 1)
	require.NoError(t, err)
	require.Len(t, limited, 1)

	require.NoError(t, store.DeleteKnowledge(ctx, "wave48-delete-fts"))
	_, err = store.GetKnowledgeHistory(ctx, "wave48-delete-fts")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))

	var count int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM knowledge_fts WHERE source_id = ?`,
		"wave48-delete-fts",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	got, err := store.SearchKnowledge(ctx, "wave48limit", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "wave48-keep-fts", got[0].Key)

	err = store.DeleteKnowledge(ctx, "wave48-delete-fts")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))
}

func TestWave48LearningEmptyQueryWithFTSUsesLIKEListBranches(t *testing.T) {
	store, _ := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "  wave48 trimmed timeout  ",
		ErrorPattern: "context deadline exceeded",
		Fix:          "retry with timeout",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave48 permission",
		ErrorPattern: "permission denied",
		Fix:          "check permissions",
		Category:     entlearning.CategoryPermission,
	}))
	ids := learningIDsByTrigger(t, store, ctx)
	_, err := store.client.Learning.UpdateOneID(ids["wave48 trimmed timeout"]).
		SetConfidence(0.8).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.UpdateOneID(ids["wave48 permission"]).
		SetConfidence(0.9).
		Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchLearnings(ctx, "", string(entlearning.CategoryTimeout), 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "wave48 trimmed timeout", got[0].Trigger)
	require.Equal(t, entlearning.CategoryTimeout, got[0].Category)

	listed, total, err := store.ListLearnings(ctx, "", 0, time.Time{}, 1, 5)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Empty(t, listed)
}

func TestWave48LearningBulkDeleteWithFTSStaleRowsResolvesEmpty(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	ctx := context.Background()
	cutoff := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave48 stale bulk",
		ErrorPattern: "wave48stale bulk delete",
		Fix:          "remove stale learning",
		Category:     entlearning.CategoryToolError,
	}))
	ids := learningIDsByTrigger(t, store, ctx)
	id := ids["wave48 stale bulk"]
	_, err := store.client.Learning.UpdateOneID(id).
		SetConfidence(0.1).
		Save(ctx)
	require.NoError(t, err)
	_, err = rawDB.ExecContext(
		ctx,
		`UPDATE learnings SET created_at = ? WHERE id = ?`,
		cutoff.Add(-time.Hour),
		id,
	)
	require.NoError(t, err)

	deleted, err := store.DeleteLearningsWhere(
		ctx,
		string(entlearning.CategoryToolError),
		0.5,
		cutoff,
	)
	require.NoError(t, err)
	require.Equal(t, 1, deleted)

	// Seed a stale FTS row explicitly so this test does not depend on whether
	// bulk deletes clean the index before searching.
	_, err = rawDB.ExecContext(ctx, `DELETE FROM learning_fts WHERE source_id = ?`, id.String())
	require.NoError(t, err)
	require.NoError(t, store.syncLearningFTS5WithExec(
		ctx,
		rawDB,
		id.String(),
		"wave48 stale bulk",
		"wave48stale bulk delete",
		"remove stale learning",
		false,
	))

	var ftsRows int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM learning_fts WHERE source_id = ?`,
		id.String(),
	).Scan(&ftsRows)
	require.NoError(t, err)
	require.Equal(t, 1, ftsRows)

	got, err := store.SearchLearnings(ctx, "wave48stale", "", 10)
	require.NoError(t, err)
	require.Empty(t, got)
}

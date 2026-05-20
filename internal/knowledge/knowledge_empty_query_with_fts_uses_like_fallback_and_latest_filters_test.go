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

func TestKnowledgeEmptyQueryWithFTSUsesLIKEFallbackAndLatestFilters(t *testing.T) {
	store, _ := newSaveKnowledgeAtomicRollsBackWhenFts5InsertFailsFTS5TestStore(t)
	ctx := context.Background()

	for i := 0; i < 12; i++ {
		require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
			Key:      fmt.Sprintf("knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-fact-%02d", i),
			Category: entknowledge.CategoryFact,
			Content:  fmt.Sprintf("knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 fact content %02d", i),
		}))
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-versioned-category",
		Category: entknowledge.CategoryFact,
		Content:  "old fact content should not be latest",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-versioned-category",
		Category: entknowledge.CategoryRule,
		Content:  "current rule content",
	}))

	got, err := store.SearchKnowledge(ctx, "", string(entknowledge.CategoryFact), 0)
	require.NoError(t, err)
	require.Len(t, got, 10)
	for _, entry := range got {
		require.Equal(t, entknowledge.CategoryFact, entry.Category)
		require.NotEqual(t, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-versioned-category", entry.Key)
	}

	_, err = store.GetKnowledge(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-missing")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))
}

func TestKnowledgeScoredEmptyQueryWithFTSUsesLIKEScores(t *testing.T) {
	store, _ := newSaveKnowledgeAtomicRollsBackWhenFts5InsertFailsFTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-low",
		Category: entknowledge.CategoryFact,
		Content:  "scored empty query low",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-high",
		Category: entknowledge.CategoryFact,
		Content:  "scored empty query high",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-filtered",
		Category: entknowledge.CategoryRule,
		Content:  "scored empty query filtered",
	}))
	_, err := store.client.Knowledge.Update().
		Where(entknowledge.Key("knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-high"), entknowledge.IsLatest(true)).
		SetRelevanceScore(7.5).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Knowledge.Update().
		Where(entknowledge.Key("knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-low"), entknowledge.IsLatest(true)).
		SetRelevanceScore(2.25).
		Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchKnowledgeScored(ctx, "", string(entknowledge.CategoryFact), 2)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-high", got[0].Entry.Key)
	require.Equal(t, 7.5, got[0].Score)
	require.Equal(t, "like", got[0].SearchSource)
	require.Equal(t, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-score-low", got[1].Entry.Key)
	require.Equal(t, 2.25, got[1].Score)
	require.Equal(t, "like", got[1].SearchSource)
}

func TestKnowledgeFTSLimitAndDeleteSync(t *testing.T) {
	store, rawDB := newSaveKnowledgeAtomicRollsBackWhenFts5InsertFailsFTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-delete-fts",
		Category: entknowledge.CategoryFact,
		Content:  "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilterslimit token first version",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-delete-fts",
		Category: entknowledge.CategoryFact,
		Content:  "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilterslimit token current version",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-keep-fts",
		Category: entknowledge.CategoryFact,
		Content:  "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilterslimit token sibling",
	}))

	limited, err := store.SearchKnowledge(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilterslimit", "", 1)
	require.NoError(t, err)
	require.Len(t, limited, 1)

	require.NoError(t, store.DeleteKnowledge(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-delete-fts"))
	_, err = store.GetKnowledgeHistory(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-delete-fts")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))

	var count int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM knowledge_fts WHERE source_id = ?`,
		"knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-delete-fts",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	got, err := store.SearchKnowledge(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilterslimit", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-keep-fts", got[0].Key)

	err = store.DeleteKnowledge(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8-delete-fts")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))
}

func TestLearningEmptyQueryWithFTSUsesLIKEListBranches(t *testing.T) {
	store, _ := newSaveKnowledgeAtomicRollsBackWhenFts5InsertFailsFTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "  knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 trimmed timeout  ",
		ErrorPattern: "context deadline exceeded",
		Fix:          "retry with timeout",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 permission",
		ErrorPattern: "permission denied",
		Fix:          "check permissions",
		Category:     entlearning.CategoryPermission,
	}))
	ids := learningIDsByTrigger(t, store, ctx)
	_, err := store.client.Learning.UpdateOneID(ids["knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 trimmed timeout"]).
		SetConfidence(0.8).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.UpdateOneID(ids["knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 permission"]).
		SetConfidence(0.9).
		Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchLearnings(ctx, "", string(entlearning.CategoryTimeout), 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 trimmed timeout", got[0].Trigger)
	require.Equal(t, entlearning.CategoryTimeout, got[0].Category)

	listed, total, err := store.ListLearnings(ctx, "", 0, time.Time{}, 1, 5)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Empty(t, listed)
}

func TestLearningBulkDeleteWithFTSStaleRowsResolvesEmpty(t *testing.T) {
	store, rawDB := newSaveKnowledgeAtomicRollsBackWhenFts5InsertFailsFTS5TestStore(t)
	ctx := context.Background()
	cutoff := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 stale bulk",
		ErrorPattern: "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFiltersstale bulk delete",
		Fix:          "remove stale learning",
		Category:     entlearning.CategoryToolError,
	}))
	ids := learningIDsByTrigger(t, store, ctx)
	id := ids["knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 stale bulk"]
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
		"knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFilters8 stale bulk",
		"knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFiltersstale bulk delete",
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

	got, err := store.SearchLearnings(ctx, "knowledgeEmptyQueryWithFtsUsesLikeFallbackAndLatestFiltersstale", "", 10)
	require.NoError(t, err)
	require.Empty(t, got)
}

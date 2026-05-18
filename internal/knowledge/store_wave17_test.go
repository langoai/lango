package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/enttest"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/search"
	_ "github.com/mattn/go-sqlite3"
)

func TestWave17SetFTS5IndexSettersStoreAndClearIndexes(t *testing.T) {
	store := newTestStore(t)

	knowledgeIdx := search.NewFTS5Index(nil, "knowledge_wave17_fts", []string{"key", "content"})
	learningIdx := search.NewFTS5Index(
		nil,
		"learning_wave17_fts",
		[]string{"trigger", "error_pattern", "fix"},
	)

	store.SetFTS5Index(knowledgeIdx)
	store.SetLearningFTS5Index(learningIdx)
	require.Same(t, knowledgeIdx, store.fts5Index)
	require.Same(t, learningIdx, store.learningFTS5Idx)

	store.SetFTS5Index(nil)
	store.SetLearningFTS5Index(nil)
	require.Nil(t, store.fts5Index)
	require.Nil(t, store.learningFTS5Idx)
}

func TestWave17SearchKnowledgeScoredFallsBackWhenFTS5SearchErrors(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:wave17-scored-fallback?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	rawDB, err := sql.Open("sqlite3", "file:wave17-scored-fallback-fts?mode=memory&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { rawDB.Close() })
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	store := NewStore(client, zap.NewNop().Sugar())
	ctx := context.Background()
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "scored-fallback",
		Category: entknowledge.CategoryFact,
		Content:  "fallback scored needle",
	}))
	_, err = store.client.Knowledge.Update().
		Where(entknowledge.Key("scored-fallback")).
		SetRelevanceScore(4.25).
		Save(ctx)
	require.NoError(t, err)

	brokenIdx := search.NewFTS5Index(rawDB, "broken_scored_fts", []string{"key", "content"})
	require.NoError(t, brokenIdx.EnsureTable())
	store.SetFTS5Index(brokenIdx)
	require.NoError(t, brokenIdx.DropTable())

	got, err := store.SearchKnowledgeScored(ctx, "needle", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "scored-fallback", got[0].Entry.Key)
	require.Equal(t, 4.25, got[0].Score)
	require.Equal(t, "like", got[0].SearchSource)
}

func TestWave17SearchLearningsFallsBackWhenFTS5SearchErrors(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:wave17-learning-fallback?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	rawDB, err := sql.Open("sqlite3", "file:wave17-learning-fallback-fts?mode=memory&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { rawDB.Close() })
	if !search.ProbeFTS5(rawDB) {
		t.Skip("FTS5 not available in current SQLite runtime")
	}

	store := NewStore(client, zap.NewNop().Sugar())
	ctx := context.Background()
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "learning fallback trigger",
		ErrorPattern: "wave17 fallback needle",
		Fix:          "retry with a clean index",
		Category:     entlearning.CategoryToolError,
	}))

	brokenIdx := search.NewFTS5Index(
		rawDB,
		"broken_learning_fts",
		[]string{"trigger", "error_pattern", "fix"},
	)
	require.NoError(t, brokenIdx.EnsureTable())
	store.SetLearningFTS5Index(brokenIdx)
	require.NoError(t, brokenIdx.DropTable())

	got, err := store.SearchLearnings(ctx, "needle", "", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "learning fallback trigger", got[0].Trigger)
	require.Equal(t, "wave17 fallback needle", got[0].ErrorPattern)
}

func TestWave17SaveKnowledgeAtomicProtectedUpdateAndDedupSyncsLatestFTS5(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	entry := KnowledgeEntry{
		Key:         "wave17-atomic-knowledge",
		Category:    entknowledge.CategoryFact,
		Content:     "first protected content with alice@example.com",
		Source:      "tool:save_knowledge",
		SourceClass: "private-confidential",
		AssetLabel:  "knowledge/wave17-atomic",
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	entry.Content = "second protected content with token SECRETSECRETSECRETSECRETSECRETSECRET"
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	history, err := store.GetKnowledgeHistory(ctx, entry.Key)
	require.NoError(t, err)
	require.Len(t, history, 2)
	require.Equal(t, 2, history[0].Version)
	require.Equal(t, entry.Content, history[0].Content)

	rows, err := store.client.Knowledge.Query().
		Where(entknowledge.Key(entry.Key)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	for _, row := range rows {
		require.NotNil(t, row.ContentCiphertext)
		require.NotContains(t, row.Content, "alice@example.com")
		require.NotContains(t, row.Content, "SECRETSECRETSECRETSECRETSECRETSECRET")
	}

	var latestCount int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM knowledge_fts WHERE source_id = ? AND knowledge_fts MATCH 'second'`,
		entry.Key,
	).Scan(&latestCount)
	require.NoError(t, err)
	require.Equal(t, 1, latestCount)

	var oldCount int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM knowledge_fts WHERE source_id = ? AND knowledge_fts MATCH 'first'`,
		entry.Key,
	).Scan(&oldCount)
	require.NoError(t, err)
	require.Equal(t, 0, oldCount)
}

func TestWave17SaveKnowledgeFTS5SyncFailuresDoNotBlockEntWrites(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()
	require.NoError(t, store.fts5Index.DropTable())

	entry := KnowledgeEntry{
		Key:      "wave17-sync-failure",
		Category: entknowledge.CategoryFact,
		Content:  "first content survives failed FTS insert",
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	entry.Content = "second content survives failed FTS update"
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	got, err := store.GetKnowledge(ctx, entry.Key)
	require.NoError(t, err)
	require.Equal(t, 2, got.Version)
	require.Equal(t, entry.Content, got.Content)

	require.NoError(t, store.DeleteKnowledge(ctx, entry.Key))
	_, err = store.GetKnowledge(ctx, entry.Key)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))

	var tableName string
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'knowledge_fts'`,
	).Scan(&tableName)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestWave17SaveLearningSyncAndDeleteFailuresDoNotBlockEntWrites(t *testing.T) {
	store, rawDB := newFTS5TestStore(t)
	ctx := context.Background()
	require.NoError(t, store.learningFTS5Idx.DropTable())

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave17 learning sync failure",
		ErrorPattern: "learning FTS insert failure",
		Fix:          "Ent row still persists",
		Category:     entlearning.CategoryToolError,
	}))
	id := learningIDsByTrigger(t, store, ctx)["wave17 learning sync failure"]

	got, err := store.GetLearning(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "learning FTS insert failure", got.ErrorPattern)

	require.NoError(t, store.DeleteLearning(ctx, id))
	_, err = store.GetLearning(ctx, id)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLearningNotFound))

	var tableName string
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'learning_fts'`,
	).Scan(&tableName)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestWave17KnowledgeVersionEquivalentReportsSourceClassErrors(t *testing.T) {
	validExisting := &ent.Knowledge{
		Key:         "source-class-entry",
		Category:    entknowledge.CategoryFact,
		Content:     "same content",
		Source:      "tool:save_knowledge",
		SourceClass: "private-confidential",
		AssetLabel:  "knowledge/source-class",
	}

	same, err := knowledgeVersionEquivalent(validExisting, KnowledgeEntry{
		Key:         "source-class-entry",
		Category:    entknowledge.CategoryFact,
		Content:     "same content",
		Source:      "tool:save_knowledge",
		SourceClass: "partner-private",
		AssetLabel:  "knowledge/source-class",
	}, "same content")
	require.Error(t, err)
	require.False(t, same)
	require.Contains(t, err.Error(), "invalid knowledge source class")

	invalidExisting := *validExisting
	invalidExisting.SourceClass = "legacy-secret"
	same, err = knowledgeVersionEquivalent(&invalidExisting, KnowledgeEntry{
		Key:         "source-class-entry",
		Category:    entknowledge.CategoryFact,
		Content:     "same content",
		Source:      "tool:save_knowledge",
		SourceClass: "private-confidential",
		AssetLabel:  "knowledge/source-class",
	}, "same content")
	require.Error(t, err)
	require.False(t, same)
	require.Contains(t, err.Error(), "invalid stored source class")
}

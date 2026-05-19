package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/eventbus"
)

func TestWave51SaveKnowledgeRetriesUniqueConflictAndPreservesRows(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.client.Knowledge.Create().
		SetKey("wave51-conflict").
		SetCategory(entknowledge.CategoryFact).
		SetContent("stale non-latest row").
		SetVersion(1).
		SetIsLatest(false).
		Save(ctx)
	require.NoError(t, err)

	err = store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave51-conflict",
		Category: entknowledge.CategoryFact,
		Content:  "new row cannot reuse version one",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "UNIQUE constraint failed")

	rows, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("wave51-conflict")).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.False(t, rows[0].IsLatest)
	require.Equal(t, "stale non-latest row", rows[0].Content)
}

func TestWave51SaveKnowledgeAtomicPersistsProtectedUpdateAndEvents(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	bus := eventbus.New()
	store.SetEventBus(bus)
	var events []eventbus.ContentSavedEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.ContentSavedEvent) {
		events = append(events, evt)
	})

	entry := KnowledgeEntry{
		Key:         "wave51-atomic-knowledge",
		Category:    entknowledge.CategoryFact,
		Content:     "wave51 first protected content for alice@example.com",
		Tags:        []string{"wave51", "atomic"},
		Source:      "tool:wave51",
		SourceClass: "private-confidential",
		AssetLabel:  "knowledge/wave51",
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	require.NoError(t, store.IncrementKnowledgeUseCount(ctx, entry.Key))
	require.NoError(t, store.BoostRelevanceScore(ctx, entry.Key, 2.5, 5.0))

	entry.Content = "wave51 second protected content with token SECRETSECRETSECRETSECRETSECRETSECRET"
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))

	got, err := store.GetKnowledge(ctx, entry.Key)
	require.NoError(t, err)
	require.Equal(t, 2, got.Version)
	require.Equal(t, entry.Content, got.Content)
	require.Equal(t, entry.Tags, got.Tags)
	require.Equal(t, entry.Source, got.Source)
	require.Equal(t, entry.SourceClass, got.SourceClass)
	require.Equal(t, entry.AssetLabel, got.AssetLabel)

	rows, err := store.client.Knowledge.Query().
		Where(entknowledge.Key(entry.Key)).
		Order(entknowledge.ByVersion()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.False(t, rows[0].IsLatest)
	require.True(t, rows[1].IsLatest)
	require.Equal(t, 1, rows[1].UseCount)
	require.Equal(t, 3.5, rows[1].RelevanceScore)
	require.NotNil(t, rows[1].ContentCiphertext)
	require.NotContains(t, rows[1].Content, "SECRETSECRETSECRETSECRETSECRETSECRET")

	var currentFTSRows int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM knowledge_fts WHERE source_id = ? AND knowledge_fts MATCH 'second'`,
		entry.Key,
	).Scan(&currentFTSRows)
	require.NoError(t, err)
	require.Equal(t, 1, currentFTSRows)

	var oldFTSRows int
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT count(*) FROM knowledge_fts WHERE source_id = ? AND knowledge_fts MATCH 'first'`,
		entry.Key,
	).Scan(&oldFTSRows)
	require.NoError(t, err)
	require.Equal(t, 0, oldFTSRows)

	require.Len(t, events, 2)
	require.Equal(t, entry.Key, events[0].ID)
	require.Equal(t, "knowledge", events[0].Collection)
	require.Equal(t, map[string]string{"category": string(entknowledge.CategoryFact)}, events[0].Metadata)
	require.Equal(t, "knowledge", events[0].Source)
	require.True(t, events[0].IsNew)
	require.True(t, events[0].NeedsGraph)
	require.Equal(t, 1, events[0].Version)
	require.Contains(t, events[0].Content, "wave51 first protected content")
	require.NotContains(t, events[0].Content, "alice@example.com")
	require.Equal(t, entry.Key, events[1].ID)
	require.Equal(t, "knowledge", events[1].Collection)
	require.False(t, events[1].IsNew)
	require.False(t, events[1].NeedsGraph)
	require.Equal(t, 2, events[1].Version)
	require.NotContains(t, events[1].Content, "SECRETSECRETSECRETSECRETSECRETSECRET")
}

func TestWave51RelevanceUpdatesAffectLatestRowsAndSurfaceBackendErrors(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave51-relevance-versioned",
		Category: entknowledge.CategoryFact,
		Content:  "old content",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave51-relevance-versioned",
		Category: entknowledge.CategoryFact,
		Content:  "new content",
	}))
	rows, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("wave51-relevance-versioned")).
		Order(entknowledge.ByVersion()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	_, err = store.client.Knowledge.UpdateOneID(rows[0].ID).SetRelevanceScore(0.2).Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Knowledge.UpdateOneID(rows[1].ID).SetRelevanceScore(4.8).Save(ctx)
	require.NoError(t, err)

	require.NoError(t, store.BoostRelevanceScore(ctx, "wave51-relevance-versioned", 0.5, 5.0))

	refreshed, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("wave51-relevance-versioned")).
		Order(entknowledge.ByVersion()).
		All(ctx)
	require.NoError(t, err)
	require.Equal(t, 0.2, refreshed[0].RelevanceScore)
	require.Equal(t, 5.0, refreshed[1].RelevanceScore)

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave51-relevance-low",
		Category: entknowledge.CategoryFact,
		Content:  "low score",
	}))
	_, err = store.client.Knowledge.Update().
		Where(entknowledge.Key("wave51-relevance-low"), entknowledge.IsLatest(true)).
		SetRelevanceScore(0.3).
		Save(ctx)
	require.NoError(t, err)

	updated, err := store.DecayAllRelevanceScores(ctx, 0.5, 0.25)
	require.NoError(t, err)
	require.Equal(t, 2, updated)

	latestVersioned, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("wave51-relevance-versioned"), entknowledge.IsLatest(true)).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, 4.5, latestVersioned.RelevanceScore)
	low, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("wave51-relevance-low"), entknowledge.IsLatest(true)).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, 0.25, low.RelevanceScore)

	closedStore := newTestStore(t)
	require.NoError(t, closedStore.client.Close())
	err = closedStore.BoostRelevanceScore(ctx, "missing", 0.5, 5.0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "boost relevance score")
	_, err = closedStore.DecayAllRelevanceScores(ctx, 0.5, 0.25)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decay relevance scores")
}

func TestWave51FTSDeleteCleanupFailuresDoNotBlockEntityDeletes(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "wave51-delete-knowledge-fts",
		Category: entknowledge.CategoryFact,
		Content:  "wave51 delete knowledge fts cleanup",
	}))
	require.NoError(t, store.fts5Index.DropTable())
	require.NoError(t, store.DeleteKnowledge(ctx, "wave51-delete-knowledge-fts"))
	_, err := store.GetKnowledge(ctx, "wave51-delete-knowledge-fts")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))
	assertMissingSQLiteTable(t, rawDB, "knowledge_fts")

	require.NoError(t, recreateWave51LearningIndex(store))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave51 delete learning fts",
		ErrorPattern: "wave51 learning cleanup",
		Fix:          "delete still succeeds",
		Category:     entlearning.CategoryToolError,
	}))
	ids := learningIDsByTrigger(t, store, ctx)
	id := ids["wave51 delete learning fts"]
	require.NoError(t, store.learningFTS5Idx.DropTable())
	require.NoError(t, store.DeleteLearning(ctx, id))
	_, err = store.GetLearning(ctx, id)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLearningNotFound))
	assertMissingSQLiteTable(t, rawDB, "learning_fts")
}

func TestWave51SaveLearningAtomicPersistsProtectedPayloadFTSAndEvent(t *testing.T) {
	store, rawDB := newWave21FTS5TestStore(t)
	store.SetPayloadProtector(stubPayloadProtector{})
	ctx := context.Background()

	bus := eventbus.New()
	store.SetEventBus(bus)
	var events []eventbus.ContentSavedEvent
	eventbus.SubscribeTyped(bus, func(evt eventbus.ContentSavedEvent) {
		events = append(events, evt)
	})

	err := store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "  wave51 protected learning  ",
		ErrorPattern: "wave51 failure for bob@example.com",
		Diagnosis:    "diagnose protected learning",
		Fix:          "rotate token SECRETSECRETSECRETSECRETSECRETSECRET",
		Category:     entlearning.CategoryToolError,
		Tags:         []string{"wave51", "learning"},
	})
	require.NoError(t, err)

	ids := learningIDsByTrigger(t, store, ctx)
	id := ids["wave51 protected learning"]
	got, err := store.GetLearning(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "wave51 protected learning", got.Trigger)
	require.Equal(t, "wave51 failure for bob@example.com", got.ErrorPattern)
	require.Equal(t, "diagnose protected learning", got.Diagnosis)
	require.Equal(t, "rotate token SECRETSECRETSECRETSECRETSECRETSECRET", got.Fix)
	require.Equal(t, []string{"wave51", "learning"}, got.Tags)

	row, err := store.client.Learning.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, row.PayloadCiphertext)
	require.NotContains(t, row.ErrorPattern, "bob@example.com")
	require.NotContains(t, row.Fix, "SECRETSECRETSECRETSECRETSECRETSECRET")

	var trigger, errorPattern, fix string
	err = rawDB.QueryRowContext(
		ctx,
		`SELECT trigger, error_pattern, fix FROM learning_fts WHERE source_id = ?`,
		id.String(),
	).Scan(&trigger, &errorPattern, &fix)
	require.NoError(t, err)
	require.Equal(t, "wave51 protected learning", trigger)
	require.NotContains(t, errorPattern, "bob@example.com")
	require.NotContains(t, fix, "SECRETSECRETSECRETSECRETSECRETSECRET")

	require.Len(t, events, 1)
	require.Equal(t, id.String(), events[0].ID)
	require.Equal(t, "learning", events[0].Collection)
	require.True(t, events[0].IsNew)
	require.False(t, events[0].NeedsGraph)
	require.Contains(t, events[0].Content, "wave51 protected learning")
	require.NotContains(t, events[0].Content, "SECRETSECRETSECRETSECRETSECRETSECRET")

	err = store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "wave51 invalid category",
		ErrorPattern: "invalid category should fail",
		Category:     entlearning.Category("wave51-invalid"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "create learning")
}

func recreateWave51LearningIndex(store *Store) error {
	if store.learningFTS5Idx == nil {
		return nil
	}
	return store.learningFTS5Idx.EnsureTable()
}

func assertMissingSQLiteTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	var name string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
		table,
	).Scan(&name)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestWave51ResolveLearningEntryNilIsEmpty(t *testing.T) {
	store := newTestStore(t)

	got, err := store.resolveLearningEntry(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, LearningEntry{}, got)
}

func TestWave51KeywordPredicatesIgnoreWhitespaceOnlyTerms(t *testing.T) {
	require.Empty(t, knowledgeKeywordPredicates(strings.Repeat(" ", 3)))
	require.Empty(t, learningKeywordPredicates("\t\n "))
}

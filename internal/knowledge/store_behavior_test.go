package knowledge

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
)

func TestSaveToolResult_TruncatesAndPersistsObservableKnowledge(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	longResult := strings.Repeat("x", 4100)
	err := store.SaveToolResult(ctx, "session-1", "web_search", nil, longResult)
	require.NoError(t, err)

	got, err := store.GetKnowledge(ctx, "tool_result:session-1:web_search")
	require.NoError(t, err)
	require.Equal(t, entknowledge.CategoryFact, got.Category)
	require.Equal(t, "tool:web_search", got.Source)
	require.Len(t, got.Content, 4096)
	require.Equal(t, strings.Repeat("x", 4096), got.Content)
}

func TestSearchKnowledgeScored_LIKEFallbackScoresLatestAndFiltersCategory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	entries := []KnowledgeEntry{
		{
			Key:      "deploy-fact-high",
			Category: entknowledge.CategoryFact,
			Content:  "deploy server config",
		},
		{
			Key:      "deploy-rule-low",
			Category: entknowledge.CategoryRule,
			Content:  "deploy server config",
		},
		{
			Key:      "deploy-fact-old",
			Category: entknowledge.CategoryFact,
			Content:  "old deploy instructions",
		},
	}
	for _, entry := range entries {
		require.NoError(t, store.SaveKnowledge(ctx, "session-1", entry))
	}
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "deploy-fact-old",
		Category: entknowledge.CategoryFact,
		Content:  "new archived instructions",
	}))
	_, err := store.client.Knowledge.Update().
		Where(entknowledge.Key("deploy-fact-high")).
		SetRelevanceScore(3.5).
		Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchKnowledgeScored(ctx, "server", "", 0)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "deploy-fact-high", got[0].Entry.Key)
	require.Equal(t, 3.5, got[0].Score)
	require.Equal(t, "like", got[0].SearchSource)
	require.Equal(t, "deploy-rule-low", got[1].Entry.Key)
	require.Equal(t, 1.0, got[1].Score)
	require.Equal(t, "like", got[1].SearchSource)

	filtered, err := store.SearchKnowledgeScored(
		ctx,
		"server",
		string(entknowledge.CategoryFact),
		10,
	)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "deploy-fact-high", filtered[0].Entry.Key)
}

func TestSearchRecentKnowledge_ReturnsLatestFilteredResults(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "recent-one",
		Category: entknowledge.CategoryFact,
		Content:  "alpha old content",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "recent-two",
		Category: entknowledge.CategoryFact,
		Content:  "alpha current content",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "recent-one",
		Category: entknowledge.CategoryFact,
		Content:  "beta current content",
	}))

	got, err := store.SearchRecentKnowledge(ctx, "alpha", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "recent-two", got[0].Key)
	require.Equal(t, "alpha current content", got[0].Content)

	latest, err := store.SearchRecentKnowledge(ctx, "beta", 1)
	require.NoError(t, err)
	require.Len(t, latest, 1)
	require.Equal(t, "recent-one", latest[0].Key)
	require.Equal(t, 2, latest[0].Version)
}

func TestGetKnowledgeByKeys_EdgeCases(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	got, err := store.GetKnowledgeByKeys(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, got)

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "lookup-one",
		Category: entknowledge.CategoryFact,
		Content:  "first lookup",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "lookup-two",
		Category: entknowledge.CategoryRule,
		Content:  "second lookup",
	}))

	got, err = store.GetKnowledgeByKeys(ctx, []string{"lookup-two", "lookup-one", "lookup-two"})
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "lookup-two", got[0].Key)
	require.Equal(t, "lookup-one", got[1].Key)
	require.Equal(t, "lookup-two", got[2].Key)

	got, err = store.GetKnowledgeByKeys(ctx, []string{"lookup-one", "missing-key"})
	require.Error(t, err)
	require.Nil(t, got)
	require.True(t, errors.Is(err, ErrKnowledgeNotFound))
	require.Contains(t, err.Error(), "missing-key")
}

func TestResetAllRelevanceScores_OnlyTouchesLatestVersions(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "reset-versioned",
		Category: entknowledge.CategoryFact,
		Content:  "first",
	}))
	_, err := store.client.Knowledge.Update().
		Where(entknowledge.Key("reset-versioned"), entknowledge.IsLatest(true)).
		SetRelevanceScore(9).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "reset-versioned",
		Category: entknowledge.CategoryFact,
		Content:  "second",
	}))
	require.NoError(t, store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "reset-single",
		Category: entknowledge.CategoryFact,
		Content:  "single",
	}))
	_, err = store.client.Knowledge.Update().
		Where(entknowledge.IsLatest(true)).
		SetRelevanceScore(4).
		Save(ctx)
	require.NoError(t, err)

	n, err := store.ResetAllRelevanceScores(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	latest, err := store.client.Knowledge.Query().
		Where(entknowledge.IsLatest(true)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, latest, 2)
	for _, entry := range latest {
		require.Equal(t, 1.0, entry.RelevanceScore)
	}

	old, err := store.client.Knowledge.Query().
		Where(entknowledge.Key("reset-versioned"), entknowledge.Version(1)).
		Only(ctx)
	require.NoError(t, err)
	require.Equal(t, 9.0, old.RelevanceScore)
}

func TestLearningReadAndScoredSearch_EdgeCases(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.GetLearning(ctx, uuid.New())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLearningNotFound))

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "timeout low",
		ErrorPattern: "deadline exceeded",
		Fix:          "retry with backoff",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "timeout high",
		ErrorPattern: "request timeout",
		Fix:          "increase timeout",
		Category:     entlearning.CategoryTimeout,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "permission denied",
		ErrorPattern: "permission failure",
		Fix:          "change permissions",
		Category:     entlearning.CategoryPermission,
	}))

	ids := learningIDsByTrigger(t, store, ctx)
	_, err = store.client.Learning.UpdateOneID(ids["timeout low"]).SetConfidence(0.2).Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.UpdateOneID(ids["timeout high"]).SetConfidence(0.9).Save(ctx)
	require.NoError(t, err)

	got, err := store.SearchLearningsScored(ctx, "", string(entlearning.CategoryTimeout), 0)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "timeout high", got[0].Entry.Trigger)
	require.Equal(t, 0.9, got[0].Score)
	require.Equal(t, "like", got[0].SearchSource)
	require.Equal(t, "timeout low", got[1].Entry.Trigger)
	require.Equal(t, 0.2, got[1].Score)
	require.Equal(t, "like", got[1].SearchSource)
}

func TestBoostLearningConfidence_BoostAndClamp(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "boost learning",
		ErrorPattern: "flaky tool",
		Fix:          "retry once",
		Category:     entlearning.CategoryToolError,
	}))
	require.NoError(t, store.SaveLearning(ctx, "session-1", LearningEntry{
		Trigger:      "clamp learning",
		ErrorPattern: "almost perfect",
		Fix:          "keep fix",
		Category:     entlearning.CategoryToolError,
	}))
	ids := learningIDsByTrigger(t, store, ctx)

	_, err := store.client.Learning.UpdateOneID(ids["boost learning"]).
		SetConfidence(0.4).
		SetSuccessCount(1).
		SetOccurrenceCount(2).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, store.BoostLearningConfidence(ctx, ids["boost learning"], 2, 0.25))
	boosted, err := store.client.Learning.Get(ctx, ids["boost learning"])
	require.NoError(t, err)
	require.Equal(t, 3, boosted.SuccessCount)
	require.Equal(t, 3, boosted.OccurrenceCount)
	require.Equal(t, 0.65, boosted.Confidence)

	_, err = store.client.Learning.UpdateOneID(ids["clamp learning"]).
		SetConfidence(0.95).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, store.BoostLearningConfidence(ctx, ids["clamp learning"], 1, 0.2))
	clamped, err := store.client.Learning.Get(ctx, ids["clamp learning"])
	require.NoError(t, err)
	require.Equal(t, 1.0, clamped.Confidence)

	err = store.BoostLearningConfidence(ctx, uuid.New(), 1, 0.1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "get learning")
}

func TestListAndBulkDeleteLearnings_FilterEdges(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	_, err := store.client.Learning.Create().
		SetTrigger("old high timeout").
		SetErrorPattern("timeout old").
		SetFix("increase timeout").
		SetCategory(entlearning.CategoryTimeout).
		SetConfidence(0.9).
		SetCreatedAt(now.Add(-48 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.Create().
		SetTrigger("new low timeout").
		SetErrorPattern("timeout new").
		SetFix("retry later").
		SetCategory(entlearning.CategoryTimeout).
		SetConfidence(0.2).
		SetCreatedAt(now).
		Save(ctx)
	require.NoError(t, err)
	_, err = store.client.Learning.Create().
		SetTrigger("old permission").
		SetErrorPattern("permission old").
		SetFix("chmod").
		SetCategory(entlearning.CategoryPermission).
		SetConfidence(0.1).
		SetCreatedAt(now.Add(-72 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	listed, total, err := store.ListLearnings(ctx, string(entlearning.CategoryTimeout), 0.5, now.Add(-24*time.Hour), 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, listed, 1)
	require.Equal(t, "old high timeout", listed[0].Trigger)

	_, err = store.DeleteLearningsWhere(ctx, "", 0, time.Time{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one filter criterion")

	deleted, err := store.DeleteLearningsWhere(ctx, "", 0.15, time.Time{})
	require.NoError(t, err)
	require.Equal(t, 1, deleted)

	deleted, err = store.DeleteLearningsWhere(ctx, "", 0, now.Add(-24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 1, deleted)
}

func TestExternalRefs_EmptySummaryAndEmptySearch(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, store.SaveExternalRef(ctx, "go-docs", "url", "https://go.dev/doc", ""))
	require.NoError(t, store.SaveExternalRef(ctx, "ent-docs", "url", "https://entgo.io", "graph orm"))
	require.NoError(t, store.SaveExternalRef(ctx, "ent-docs", "url", "https://entgo.io/docs", ""))

	got, err := store.SearchExternalRefs(ctx, "")
	require.NoError(t, err)
	require.Len(t, got, 2)
	byName := map[string]ExternalRefEntry{}
	for _, ref := range got {
		byName[ref.Name] = ref
	}
	require.Equal(t, "https://go.dev/doc", byName["go-docs"].Location)
	require.Empty(t, byName["go-docs"].Summary)
	require.Equal(t, "https://entgo.io/docs", byName["ent-docs"].Location)
	require.Equal(t, "graph orm", byName["ent-docs"].Summary)
}

func TestKnowledgeStore_ClosedBackendReturnsErrors(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	require.NoError(t, store.client.Close())

	_, err := store.GetKnowledge(ctx, "missing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "query knowledge")

	err = store.SaveKnowledge(ctx, "session-1", KnowledgeEntry{
		Key:      "closed-save",
		Category: entknowledge.CategoryFact,
		Content:  "cannot persist",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "query knowledge")

	_, err = store.SearchKnowledge(ctx, "anything", "", 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "search knowledge")

	_, err = store.SearchRecentKnowledge(ctx, "", 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "search recent knowledge")

	_, err = store.ResetAllRelevanceScores(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reset relevance scores")

	err = store.DeleteKnowledge(ctx, "closed-delete")
	require.Error(t, err)
	require.Contains(t, err.Error(), "delete knowledge")
	require.False(t, errors.Is(err, ErrKnowledgeNotFound))
}

package knowledge

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
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

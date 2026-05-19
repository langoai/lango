package adk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/graph"
	internal "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

type wave35ContentRetriever struct {
	query   string
	opts    graph.ContentRetrieveOptions
	results []graph.ContentResult
	err     error
}

func (r *wave35ContentRetriever) Retrieve(_ context.Context, query string, opts graph.ContentRetrieveOptions) ([]graph.ContentResult, error) {
	r.query = query
	r.opts = opts
	return r.results, r.err
}

func TestWave35RetrieveGraphRAGDataUsesDefaultLimitAndSessionKey(t *testing.T) {
	t.Parallel()

	retriever := &wave35ContentRetriever{
		results: []graph.ContentResult{{
			Collection: "session",
			SourceID:   "turn-1",
			Content:    "retrieved context",
			Score:      0.88,
		}},
	}
	graphRAG := graph.NewGraphRAGService(retriever, nil, 1, 1, zap.NewNop().Sugar())
	adapter := newTestContextAdapter(t, nil).WithGraphRAG(graphRAG, 0)

	got := adapter.retrieveGraphRAGData(context.Background(), "find context", "discord:guild:channel")

	require.NotNil(t, got)
	assert.Equal(t, "find context", retriever.query)
	assert.Equal(t, 5, retriever.opts.Limit)
	assert.Equal(t, "discord:guild:channel", retriever.opts.SessionKey)
	require.Len(t, got.ContentResults, 1)
	assert.Equal(t, "retrieved context", got.ContentResults[0].Content)
}

func TestWave35FormatGraphRAGSectionTruncatesContentBeforeGraphResults(t *testing.T) {
	t.Parallel()

	graphRAG := graph.NewGraphRAGService(nil, nil, 1, 1, zap.NewNop().Sugar())
	adapter := newTestContextAdapter(t, nil).WithGraphRAG(graphRAG, 3)
	result := &graph.GraphRAGResult{
		ContentResults: []graph.ContentResult{
			{
				Collection: "session",
				SourceID:   "small",
				Content:    "short",
			},
			{
				Collection: "session",
				SourceID:   "large",
				Content:    "this content should not fit in the tiny test budget",
			},
		},
		GraphResults: []graph.GraphNode{{
			ID:        "node-1",
			NodeType:  "fact",
			Predicate: graph.RelatedTo,
			FromNode:  "session:small",
		}},
	}

	got := adapter.formatGraphRAGSection(result, 6)

	assert.Contains(t, got, "## Retrieved Context")
	assert.Contains(t, got, "short")
	assert.NotContains(t, got, "large")
	assert.NotContains(t, got, "Graph-Expanded Context")
	assert.Len(t, result.ContentResults, 1, "formatting mutates the result to the fitted prefix")
	assert.Empty(t, result.GraphResults)
}

func TestWave35FormatRecallSectionAppliesBudgetAndSkipsEmptySummaries(t *testing.T) {
	t.Parallel()

	matches := []RecallMatch{
		{SessionKey: "older-session", Summary: "use the cached SSH key", Rank: 0.97},
		{SessionKey: "empty-session", Summary: "", Rank: 0.80},
		{SessionKey: "overflow-session", Summary: "this lower-ranked summary should not fit", Rank: 0.51},
	}

	unlimited := formatRecallSection(matches, 0)
	assert.Contains(t, unlimited, "## Prior Session Recall")
	assert.Contains(t, unlimited, "### [older-session] (rank=0.97)")
	assert.Contains(t, unlimited, "use the cached SSH key")
	assert.NotContains(t, unlimited, "### [empty-session]")

	budgeted := formatRecallSection(matches, 24)
	assert.Contains(t, budgeted, "older-session")
	assert.NotContains(t, budgeted, "overflow-session")

	assert.Empty(t, formatRecallSection(matches, 1))
	assert.Empty(t, formatRecallSection(nil, 100))
}

func TestWave35DiscardActiveChildWithoutReasonRollsBackOverlayWithoutParentNote(t *testing.T) {
	t.Parallel()

	store := newMockStore()
	sess := &internal.Session{
		Key:       "parent-session",
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, store.Create(sess))

	var lifecycle []string
	adapter := NewSessionAdapter(sess, store, "lango-orchestrator")
	svc := NewSessionServiceAdapter(store, "lango-orchestrator").
		WithChildLifecycleHook(func(ev internal.SessionLifecycleEvent) {
			lifecycle = append(lifecycle, ev.Type)
		}).
		WithIsolatedAgents([]string{"operator"})

	require.NoError(t, svc.AppendEvent(context.Background(), adapter, newTestEvent("operator", "model", "isolated reply")))
	require.Len(t, adapter.sess.History, 1, "isolated child writes are overlaid until merge/discard")

	require.NoError(t, svc.DiscardActiveChild("parent-session"))

	assert.Equal(t, []string{"fork", "discard"}, lifecycle)
	assert.Empty(t, store.messages["parent-session"], "discard without reason must not persist a compact note")
	assert.Empty(t, adapter.sess.History, "discard should roll back the temporary parent overlay")
	assert.Nil(t, svc.activeChild["parent-session"])
	children, err := svc.childStore.ChildrenOf("parent-session")
	require.NoError(t, err)
	assert.Empty(t, children)
}

func TestWave35CleanupFailedTurnWithNilStoreIsNoop(t *testing.T) {
	t.Parallel()

	svc := NewSessionServiceAdapter(nil, "lango-orchestrator")

	require.NoError(t, svc.CleanupFailedTurn("missing-session", "agent error"))
	require.NoError(t, svc.DiscardActiveChild("missing-session"))
}

func TestWave35DanglingToolCallsNormalizesEmptyIDsAndMatchesResponsesByName(t *testing.T) {
	t.Parallel()

	history := []internal.Message{
		{
			Role:   types.RoleAssistant,
			Author: "operator",
			ToolCalls: []internal.ToolCall{{
				Name:  "lookup_context",
				Input: `{"q":"coverage"}`,
			}},
		},
		{
			Role:   types.RoleTool,
			Author: "operator",
			ToolCalls: []internal.ToolCall{{
				Name:   "lookup_context",
				Output: `{"ok":true}`,
			}},
		},
		{
			Role:   types.RoleModel,
			Author: "",
			ToolCalls: []internal.ToolCall{{
				Name:  "slow_tool",
				Input: `{"id":1}`,
			}},
		},
	}

	got := danglingToolCalls(history)

	require.Len(t, got, 1)
	assert.Equal(t, "call_slow_tool", got[0].ID)
	assert.Equal(t, "slow_tool", got[0].Name)
	assert.Empty(t, got[0].OriginAuthor)
}

func TestWave35RetrieveGraphRAGDataTreatsRetrieverErrorsAsEmptyResults(t *testing.T) {
	t.Parallel()

	retriever := &wave35ContentRetriever{err: errors.New("fts unavailable")}
	graphRAG := graph.NewGraphRAGService(retriever, nil, 1, 1, zap.NewNop().Sugar())
	adapter := newTestContextAdapter(t, nil).WithGraphRAG(graphRAG, 2)

	got := adapter.retrieveGraphRAGData(context.Background(), "query", "session-key")

	require.NotNil(t, got)
	assert.Empty(t, got.ContentResults)
	assert.Empty(t, got.GraphResults)
	assert.Equal(t, "query", retriever.query)
	assert.Equal(t, "session-key", retriever.opts.SessionKey)
}

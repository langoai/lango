package adk

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/types"
)

func TestWave25BuildContextInjectedItemsCoversNilAndItemFields(t *testing.T) {
	t.Parallel()

	assert.Nil(t, buildContextInjectedItems(nil))

	retrieved := &knowledge.RetrievalResult{
		Items: map[knowledge.ContextLayer][]knowledge.ContextItem{
			knowledge.LayerRuntimeContext: {
				{
					Key:      "runtime/current",
					Content:  "runtime branch content",
					Score:    0.71,
					Source:   "runtime",
					Category: "state",
				},
			},
			knowledge.LayerUserKnowledge: {
				{
					Key:      "user/preference",
					Content:  "abcdefghijkl",
					Score:    0.92,
					Source:   "store",
					Category: "preference",
				},
			},
		},
	}

	got := buildContextInjectedItems(retrieved)
	sort.Slice(got, func(i, j int) bool {
		return got[i].Key < got[j].Key
	})

	require.Equal(t, []eventbus.ContextInjectedItem{
		{
			Layer:         "runtime_context",
			Key:           "runtime/current",
			Score:         0.71,
			Source:        "runtime",
			Category:      "state",
			TokenEstimate: types.EstimateTokens("runtime branch content"),
		},
		{
			Layer:         "user_knowledge",
			Key:           "user/preference",
			Score:         0.92,
			Source:        "store",
			Category:      "preference",
			TokenEstimate: types.EstimateTokens("abcdefghijkl"),
		},
	}, got)
}

func TestWave25EstimateKnowledgeTokensCoversNilEmptyAndItemOverhead(t *testing.T) {
	t.Parallel()

	assert.Zero(t, estimateKnowledgeTokens(nil))
	assert.Zero(t, estimateKnowledgeTokens(&knowledge.RetrievalResult{}))

	retrieved := &knowledge.RetrievalResult{
		Items: map[knowledge.ContextLayer][]knowledge.ContextItem{
			knowledge.LayerToolRegistry: {
				{Content: "tool registry entry"},
				{Content: ""},
			},
			knowledge.LayerPendingInquiries: {
				{Content: "pending inquiry content"},
			},
		},
	}

	contentTokens := types.EstimateTokens("tool registry entry")
	contentTokens += types.EstimateTokens("pending inquiry content")

	got := estimateKnowledgeTokens(retrieved)
	assert.Greater(t, got, contentTokens)
}

func TestWave25EstimateRetrievedResultTokensCoversNilHeaderContentAndGraphNodes(t *testing.T) {
	t.Parallel()

	assert.Zero(t, estimateRetrievedResultTokens(nil))

	empty := &graph.GraphRAGResult{}
	assert.Equal(t, types.EstimateTokens("## Retrieved Context\n"), estimateRetrievedResultTokens(empty))

	result := &graph.GraphRAGResult{
		ContentResults: []graph.ContentResult{
			{Content: "first retrieved context"},
			{Content: "second retrieved context with more text"},
		},
		GraphResults: []graph.GraphNode{
			{ID: "node-1"},
			{ID: "node-2"},
		},
	}

	contentOnly := &graph.GraphRAGResult{
		ContentResults: result.ContentResults,
	}
	contentTokens := types.EstimateTokens("## Retrieved Context\n")
	contentTokens += types.EstimateTokens("first retrieved context")
	contentTokens += types.EstimateTokens("second retrieved context with more text")

	got := estimateRetrievedResultTokens(result)
	assert.Greater(t, got, contentTokens)
	assert.Greater(t, got, estimateRetrievedResultTokens(contentOnly))
}

func TestWave25EstimateMemoryTokensCoversEmptyReflectionsAndObservations(t *testing.T) {
	t.Parallel()

	assert.Zero(t, estimateMemoryTokens(nil, nil))
	assert.Zero(t, estimateMemoryTokens([]memory.Reflection{}, []memory.Observation{}))

	reflections := []memory.Reflection{
		{Content: "reflection one"},
		{Content: "reflection two is longer"},
	}
	observations := []memory.Observation{
		{Content: "observation one"},
	}

	contentTokens := types.EstimateTokens("reflection one")
	contentTokens += types.EstimateTokens("reflection two is longer")
	contentTokens += types.EstimateTokens("observation one")

	got := estimateMemoryTokens(reflections, observations)
	assert.Greater(t, got, contentTokens)
	assert.Greater(t, got, estimateMemoryTokens(reflections[:1], observations[:0]))
}

func TestWave25EstimateRunSummaryTokensCoversEmptyAndSummaryOverhead(t *testing.T) {
	t.Parallel()

	assert.Zero(t, estimateRunSummaryTokens(nil))
	assert.Zero(t, estimateRunSummaryTokens([]RunSummaryContext{}))

	summaries := []RunSummaryContext{
		{
			RunID:          "run-alpha",
			Goal:           "repair coverage baseline",
			Status:         "running",
			CurrentStep:    "write focused tests",
			CurrentBlocker: "none",
		},
		{
			RunID: "run-beta",
			Goal:  "",
		},
	}

	contentTokens := types.EstimateTokens("repair coverage baseline")
	contentTokens += types.EstimateTokens("run-alpha")
	contentTokens += types.EstimateTokens("run-beta")

	got := estimateRunSummaryTokens(summaries)
	assert.Greater(t, got, contentTokens)
	assert.Greater(t, got, estimateRunSummaryTokens(summaries[:1]))
}

func TestWave25MergeRetrievalResultsCoversNilIdentityAndDisjointMerge(t *testing.T) {
	t.Parallel()

	assert.Nil(t, mergeRetrievalResults(nil, nil))

	left := &knowledge.RetrievalResult{
		Items: map[knowledge.ContextLayer][]knowledge.ContextItem{
			knowledge.LayerRuntimeContext: {
				{Key: "runtime/current"},
			},
		},
		TotalItems: 99,
	}
	right := &knowledge.RetrievalResult{
		Items: map[knowledge.ContextLayer][]knowledge.ContextItem{
			knowledge.LayerUserKnowledge: {
				{Key: "user/preference"},
				{Key: "user/fact"},
			},
			knowledge.LayerAgentLearnings: {},
		},
		TotalItems: 42,
	}

	assert.Same(t, right, mergeRetrievalResults(nil, right))
	assert.Same(t, left, mergeRetrievalResults(left, nil))

	got := mergeRetrievalResults(left, right)
	require.NotNil(t, got)
	assert.NotSame(t, left, got)
	assert.NotSame(t, right, got)
	assert.Equal(t, 3, got.TotalItems)
	assert.Equal(t, []knowledge.ContextItem{{Key: "runtime/current"}}, got.Items[knowledge.LayerRuntimeContext])
	assert.Equal(t, []knowledge.ContextItem{{Key: "user/preference"}, {Key: "user/fact"}}, got.Items[knowledge.LayerUserKnowledge])
	assert.Empty(t, got.Items[knowledge.LayerAgentLearnings])
}

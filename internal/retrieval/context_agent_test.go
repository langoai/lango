package retrieval

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/knowledge"
)

// mockContextSource implements ContextSearchSource for testing.
type mockContextSource struct {
	results []embedding.RAGResult
	err     error
}

func (m *mockContextSource) Retrieve(_ context.Context, _ string, _ embedding.RetrieveOptions) ([]embedding.RAGResult, error) {
	return m.results, m.err
}

func TestContextSearchAgent_Name(t *testing.T) {
	agent := NewContextSearchAgent(&mockContextSource{}, embedding.RetrieveOptions{}, zaptest.NewLogger(t).Sugar())
	assert.Equal(t, "context-search", agent.Name())
}

func TestContextSearchAgent_Layers(t *testing.T) {
	agent := NewContextSearchAgent(&mockContextSource{}, embedding.RetrieveOptions{}, zaptest.NewLogger(t).Sugar())
	layers := agent.Layers()
	assert.Equal(t, []knowledge.ContextLayer{
		knowledge.LayerUserKnowledge,
		knowledge.LayerAgentLearnings,
	}, layers)
}

func TestContextSearchAgent_Search(t *testing.T) {
	tests := []struct {
		give        string
		giveResults []embedding.RAGResult
		giveErr     error
		wantLen     int
		wantFirst   Finding
	}{
		{
			give: "knowledge collection results",
			giveResults: []embedding.RAGResult{
				{Collection: "knowledge", SourceID: "deploy-config", Content: "Deploy configuration guide", Distance: 0.3},
			},
			wantLen: 1,
			wantFirst: Finding{
				Key:          "deploy-config",
				Content:      "Deploy configuration guide",
				Score:        0.7, // 1.0 - 0.3
				SearchSource: "vector",
				Agent:        "context-search",
				Layer:        knowledge.LayerUserKnowledge,
			},
		},
		{
			give: "learning collection results",
			giveResults: []embedding.RAGResult{
				{Collection: "learning", SourceID: "timeout-fix", Content: "Fix timeout by increasing deadline", Distance: 0.1},
			},
			wantLen: 1,
			wantFirst: Finding{
				Key:          "timeout-fix",
				Content:      "Fix timeout by increasing deadline",
				Score:        0.9, // 1.0 - 0.1
				SearchSource: "vector",
				Agent:        "context-search",
				Layer:        knowledge.LayerAgentLearnings,
			},
		},
		{
			give: "observation/reflection collections filtered out in v1",
			giveResults: []embedding.RAGResult{
				{Collection: "knowledge", SourceID: "k1", Content: "kept", Distance: 0.5},
				{Collection: "observation", SourceID: "obs1", Content: "filtered", Distance: 0.2},
				{Collection: "reflection", SourceID: "ref1", Content: "filtered", Distance: 0.1},
			},
			wantLen: 1,
			wantFirst: Finding{
				Key:          "k1",
				Content:      "kept",
				Score:        0.5,
				SearchSource: "vector",
				Agent:        "context-search",
				Layer:        knowledge.LayerUserKnowledge,
			},
		},
		{
			give:        "empty results",
			giveResults: nil,
			wantLen:     0,
		},
		{
			give: "perfect match (distance 0)",
			giveResults: []embedding.RAGResult{
				{Collection: "knowledge", SourceID: "exact", Content: "exact match", Distance: 0.0},
			},
			wantLen: 1,
			wantFirst: Finding{
				Key:   "exact",
				Score: 1.0,
				Layer: knowledge.LayerUserKnowledge,
			},
		},
		{
			give: "very distant (score floored at 0)",
			giveResults: []embedding.RAGResult{
				{Collection: "knowledge", SourceID: "far", Content: "far away", Distance: 1.5},
			},
			wantLen: 1,
			wantFirst: Finding{
				Key:   "far",
				Score: 0.0, // max(0, 1.0 - 1.5) = 0
				Layer: knowledge.LayerUserKnowledge,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			source := &mockContextSource{results: tt.giveResults, err: tt.giveErr}
			agent := NewContextSearchAgent(source, embedding.RetrieveOptions{}, zaptest.NewLogger(t).Sugar())

			findings, err := agent.Search(context.Background(), "test query", 10)
			require.NoError(t, err)
			assert.Len(t, findings, tt.wantLen)

			if tt.wantLen > 0 {
				f := findings[0]
				assert.Equal(t, tt.wantFirst.Key, f.Key)
				assert.InDelta(t, tt.wantFirst.Score, f.Score, 0.001)
				assert.Equal(t, tt.wantFirst.Layer, f.Layer)
				assert.Equal(t, "vector", f.SearchSource)
				assert.Equal(t, "context-search", f.Agent)
			}
		})
	}
}

func TestContextSearchAgent_Search_Error(t *testing.T) {
	source := &mockContextSource{err: assert.AnError}
	agent := NewContextSearchAgent(source, embedding.RetrieveOptions{}, zaptest.NewLogger(t).Sugar())

	_, err := agent.Search(context.Background(), "query", 10)
	assert.Error(t, err)
}

func TestCollectionToLayer(t *testing.T) {
	tests := []struct {
		give      string
		wantLayer knowledge.ContextLayer
		wantOK    bool
	}{
		{"knowledge", knowledge.LayerUserKnowledge, true},
		{"learning", knowledge.LayerAgentLearnings, true},
		{"observation", 0, false},
		{"reflection", 0, false},
		{"unknown", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			layer, ok := collectionToLayer(tt.give)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantLayer, layer)
			}
		})
	}
}

func TestVectorDistanceToScore(t *testing.T) {
	tests := []struct {
		give float32
		want float64
	}{
		{0.0, 1.0},
		{0.5, 0.5},
		{1.0, 0.0},
		{1.5, 0.0}, // floored at 0
		{0.3, 0.7},
	}
	for _, tt := range tests {
		score := vectorDistanceToScore(tt.give)
		assert.InDelta(t, tt.want, score, 0.001)
	}
}

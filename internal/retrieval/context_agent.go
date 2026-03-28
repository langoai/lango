package retrieval

import (
	"context"
	"math"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/embedding"
	"github.com/langoai/lango/internal/knowledge"
)

// ContextSearchSource provides vector/semantic search results.
// Satisfied by *embedding.RAGService.
type ContextSearchSource interface {
	Retrieve(ctx context.Context, query string, opts embedding.RetrieveOptions) ([]embedding.RAGResult, error)
}

// ContextSearchAgent performs semantic/vector search via RAGService and converts
// results to Findings. It covers the same factual layers as FactSearchAgent
// (UserKnowledge, AgentLearnings) but uses vector similarity instead of keyword
// matching. Results naturally rank below FTS5 scores (0-1 vs 1-10+), making
// this agent a "related context expansion" source, not authoritative truth.
//
// v1 scope: knowledge + learning collections only; observation/reflection
// deferred to avoid budget boundary violation with memory section.
type ContextSearchAgent struct {
	source ContextSearchSource
	opts   embedding.RetrieveOptions
	logger *zap.SugaredLogger
}

// NewContextSearchAgent creates a new ContextSearchAgent.
func NewContextSearchAgent(source ContextSearchSource, opts embedding.RetrieveOptions, logger *zap.SugaredLogger) *ContextSearchAgent {
	return &ContextSearchAgent{
		source: source,
		opts:   opts,
		logger: logger,
	}
}

// Name returns the agent identifier.
func (a *ContextSearchAgent) Name() string { return "context-search" }

// Layers returns the context layers this agent produces.
func (a *ContextSearchAgent) Layers() []knowledge.ContextLayer {
	return []knowledge.ContextLayer{
		knowledge.LayerUserKnowledge,
		knowledge.LayerAgentLearnings,
	}
}

// Search retrieves semantically similar items from vector store and converts
// to findings. Only knowledge and learning collections are included in v1.
func (a *ContextSearchAgent) Search(ctx context.Context, query string, limit int) ([]Finding, error) {
	if limit <= 0 {
		limit = 10
	}

	opts := a.opts
	opts.Limit = limit
	// v1: restrict to factual collections only.
	opts.Collections = []string{"knowledge", "learning"}

	results, err := a.source.Retrieve(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0, len(results))
	for _, r := range results {
		layer, ok := collectionToLayer(r.Collection)
		if !ok {
			continue // skip collections not mapped in v1
		}
		findings = append(findings, Finding{
			Key:          r.SourceID,
			Content:      r.Content,
			Score:        vectorDistanceToScore(r.Distance),
			SearchSource: "vector",
			Agent:        a.Name(),
			Layer:        layer,
		})
	}

	return findings, nil
}

// collectionToLayer maps RAG collection names to context layers.
// Returns false for collections not supported in v1.
func collectionToLayer(collection string) (knowledge.ContextLayer, bool) {
	switch collection {
	case "knowledge":
		return knowledge.LayerUserKnowledge, true
	case "learning":
		return knowledge.LayerAgentLearnings, true
	default:
		return 0, false
	}
}

// vectorDistanceToScore converts cosine distance (lower=better) to score (higher=better).
// Range: 0.0 (worst) to 1.0 (perfect match).
func vectorDistanceToScore(distance float32) float64 {
	return math.Max(0, 1.0-float64(distance))
}

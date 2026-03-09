package graph

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// VectorResult mirrors embedding.RAGResult to avoid an import cycle.
type VectorResult struct {
	Collection string
	SourceID   string
	Content    string
	Distance   float32
}

// VectorRetrieveOptions mirrors embedding.RetrieveOptions to avoid an import cycle.
type VectorRetrieveOptions struct {
	Collections []string
	Limit       int
	SessionKey  string
	MaxDistance float32
}

// VectorRetriever retrieves results from a vector store. Implemented by
// embedding.RAGService (satisfies via adapter).
type VectorRetriever interface {
	Retrieve(ctx context.Context, query string, opts VectorRetrieveOptions) ([]VectorResult, error)
}

// GraphRAGService provides hybrid retrieval combining vector search with graph expansion.
type GraphRAGService struct {
	vectorRAG VectorRetriever
	graph     Store
	maxDepth  int
	maxExpand int
	logger    *zap.SugaredLogger
}

// NewGraphRAGService creates a new hybrid graph RAG service.
func NewGraphRAGService(
	vectorRAG VectorRetriever,
	graph Store,
	maxDepth int,
	maxExpand int,
	logger *zap.SugaredLogger,
) *GraphRAGService {
	if maxDepth <= 0 {
		maxDepth = 2
	}
	if maxExpand <= 0 {
		maxExpand = 10
	}
	return &GraphRAGService{
		vectorRAG: vectorRAG,
		graph:     graph,
		maxDepth:  maxDepth,
		maxExpand: maxExpand,
		logger:    logger,
	}
}

// GraphRAGResult extends VectorResult with graph-expanded context.
type GraphRAGResult struct {
	// VectorResults are the original vector-search results.
	VectorResults []VectorResult
	// GraphResults are additional results discovered via graph traversal.
	GraphResults []GraphNode
}

// GraphNode represents a node discovered through graph expansion.
type GraphNode struct {
	ID        string
	Predicate string // the edge that led here
	FromNode  string // the source node this was discovered from
	Depth     int
}

// Retrieve performs 2-phase hybrid retrieval:
// Phase 1: Vector search (sqlite-vec cosine similarity)
// Phase 2: Graph expansion from Phase 1 results (depth 1-2 hops)
func (s *GraphRAGService) Retrieve(ctx context.Context, query string, opts VectorRetrieveOptions) (*GraphRAGResult, error) {
	result := &GraphRAGResult{}

	// Phase 1: Vector search via existing RAG service.
	if s.vectorRAG != nil {
		vectorResults, err := s.vectorRAG.Retrieve(ctx, query, opts)
		if err != nil {
			s.logger.Warnw("vector retrieval error", "error", err)
		} else {
			result.VectorResults = vectorResults
		}
	}

	// Phase 2: Graph expansion from vector results.
	if s.graph == nil || len(result.VectorResults) == 0 {
		return result, nil
	}

	expansionPredicates := []string{RelatedTo, ResolvedBy, CausedBy, SimilarTo}
	seen := make(map[string]bool, len(result.VectorResults))

	// Mark vector results as seen to avoid duplicates.
	for _, vr := range result.VectorResults {
		nodeID := buildNodeID(vr.Collection, vr.SourceID)
		seen[nodeID] = true
	}

	// Expand from each vector result node.
	for _, vr := range result.VectorResults {
		nodeID := buildNodeID(vr.Collection, vr.SourceID)

		triples, err := s.graph.Traverse(ctx, nodeID, s.maxDepth, expansionPredicates)
		if err != nil {
			s.logger.Debugw("graph traverse error", "node", nodeID, "error", err)
			continue
		}

		for _, t := range triples {
			target := t.Object
			if target == nodeID {
				target = t.Subject
			}
			if seen[target] {
				continue
			}
			seen[target] = true

			result.GraphResults = append(result.GraphResults, GraphNode{
				ID:        target,
				Predicate: t.Predicate,
				FromNode:  nodeID,
				Depth:     1, // simplified — real depth comes from BFS
			})

			if len(result.GraphResults) >= s.maxExpand {
				break
			}
		}

		if len(result.GraphResults) >= s.maxExpand {
			break
		}
	}

	return result, nil
}

// AssembleSection builds a formatted string section for context injection.
func (s *GraphRAGService) AssembleSection(result *GraphRAGResult) string {
	if result == nil {
		return ""
	}

	var b strings.Builder

	// Vector results section.
	if len(result.VectorResults) > 0 {
		b.WriteString("## Semantic Context (RAG)\n")
		for _, r := range result.VectorResults {
			if r.Content == "" {
				continue
			}
			fmt.Fprintf(&b, "\n### [%s] %s\n", r.Collection, r.SourceID)
			b.WriteString(r.Content)
			b.WriteString("\n")
		}
	}

	// Graph expansion section.
	if len(result.GraphResults) > 0 {
		b.WriteString("\n## Graph-Expanded Context\n")
		b.WriteString("The following related items were discovered through knowledge graph traversal:\n")
		for _, g := range result.GraphResults {
			fmt.Fprintf(&b, "- **%s** (via %s from %s)\n", g.ID, g.Predicate, g.FromNode)
		}
	}

	return b.String()
}

// buildNodeID creates a graph node identifier from collection and source ID.
func buildNodeID(collection, sourceID string) string {
	return collection + ":" + sourceID
}

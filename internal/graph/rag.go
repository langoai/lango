package graph

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ContentResult represents a single retrieved context item used to seed graph
// traversal.
type ContentResult struct {
	Collection string
	SourceID   string
	Content    string
	Score      float32
}

// ContentRetrieveOptions configures phase-1 retrieved context lookup.
type ContentRetrieveOptions struct {
	Collections []string
	Limit       int
	SessionKey  string
}

// ContentRetriever retrieves phase-1 context items from an arbitrary backend
// such as FTS5. Implementations are supplied by wiring.
type ContentRetriever interface {
	Retrieve(ctx context.Context, query string, opts ContentRetrieveOptions) ([]ContentResult, error)
}

// GraphRAGService provides hybrid retrieval combining retrieved-context seeding
// with graph expansion.
type GraphRAGService struct {
	contentRetriever ContentRetriever
	graph            Store
	maxDepth         int
	maxExpand        int
	logger           *zap.SugaredLogger
}

// NewGraphRAGService creates a new hybrid graph RAG service.
func NewGraphRAGService(
	contentRetriever ContentRetriever,
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
		contentRetriever: contentRetriever,
		graph:            graph,
		maxDepth:         maxDepth,
		maxExpand:        maxExpand,
		logger:           logger,
	}
}

// GraphRAGResult extends phase-1 content results with graph-expanded context.
type GraphRAGResult struct {
	// ContentResults are the original retrieved context results.
	ContentResults []ContentResult
	// GraphResults are additional results discovered via graph traversal.
	GraphResults []GraphNode
}

// GraphNode represents a node discovered through graph expansion.
type GraphNode struct {
	ID        string
	NodeType  string // ObjectType name from triple metadata (empty = untyped)
	Predicate string // the edge that led here
	FromNode  string // the source node this was discovered from
	Depth     int
}

// Retrieve performs 2-phase hybrid retrieval:
// Phase 1: Retrieved context lookup (e.g., FTS5/BM25)
// Phase 2: Graph expansion from phase-1 results (depth 1-2 hops)
func (s *GraphRAGService) Retrieve(ctx context.Context, query string, opts ContentRetrieveOptions) (*GraphRAGResult, error) {
	result := &GraphRAGResult{}

	// Phase 1: Retrieved context lookup via configured content retriever.
	if s.contentRetriever != nil {
		contentResults, err := s.contentRetriever.Retrieve(ctx, query, opts)
		if err != nil {
			s.logger.Warnw("content retrieval error", "error", err)
		} else {
			result.ContentResults = contentResults
		}
	}

	// Phase 2: Graph expansion from retrieved context results.
	if s.graph == nil || len(result.ContentResults) == 0 {
		return result, nil
	}

	expansionPredicates := []string{RelatedTo, ResolvedBy, CausedBy, SimilarTo}
	seen := make(map[string]bool, len(result.ContentResults))

	// Mark phase-1 results as seen to avoid duplicates.
	for _, cr := range result.ContentResults {
		nodeID := buildNodeID(cr.Collection, cr.SourceID)
		seen[nodeID] = true
	}

	// Expand from each phase-1 result node.
	for _, cr := range result.ContentResults {
		nodeID := buildNodeID(cr.Collection, cr.SourceID)

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

			nodeType := t.SubjectType
			if target == t.Object {
				nodeType = t.ObjectType
			}
			result.GraphResults = append(result.GraphResults, GraphNode{
				ID:        target,
				NodeType:  nodeType,
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

	// Retrieved context section.
	if len(result.ContentResults) > 0 {
		b.WriteString("## Retrieved Context\n")
		for _, r := range result.ContentResults {
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
			if g.NodeType != "" {
				fmt.Fprintf(&b, "- **%s:%s** (via %s from %s)\n", g.NodeType, g.ID, g.Predicate, g.FromNode)
			} else {
				fmt.Fprintf(&b, "- **%s** (via %s from %s)\n", g.ID, g.Predicate, g.FromNode)
			}
		}
	}

	return b.String()
}

// buildNodeID creates a graph node identifier from collection and source ID.
func buildNodeID(collection, sourceID string) string {
	return collection + ":" + sourceID
}

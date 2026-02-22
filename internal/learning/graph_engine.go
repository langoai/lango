package learning

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/knowledge"
)

// GraphCallback is a hook for asynchronous graph updates from the learning system.
type GraphCallback func(triples []graph.Triple)

// Compile-time interface check.
var _ ToolResultObserver = (*GraphEngine)(nil)

// GraphEngine extends the learning engine with graph-based relationship tracking
// and confidence propagation across similar learnings.
type GraphEngine struct {
	*Engine
	graphStore    graph.Store
	graphCallback GraphCallback
	propagation   float64 // confidence propagation rate (0.0-1.0)
	logger        *zap.SugaredLogger
}

// NewGraphEngine creates a graph-enhanced learning engine.
func NewGraphEngine(store *knowledge.Store, graphStore graph.Store, logger *zap.SugaredLogger) *GraphEngine {
	return &GraphEngine{
		Engine:      NewEngine(store, logger),
		graphStore:  graphStore,
		propagation: 0.3,
		logger:      logger,
	}
}

// SetGraphCallback sets the asynchronous graph update hook.
func (e *GraphEngine) SetGraphCallback(cb GraphCallback) {
	e.graphCallback = cb
}

// OnToolResult observes a tool execution result, records learnings,
// and updates the knowledge graph with error-resolution relationships.
func (e *GraphEngine) OnToolResult(ctx context.Context, sessionKey, toolName string, params map[string]interface{}, result interface{}, err error) {
	// Delegate to base engine.
	e.Engine.OnToolResult(ctx, sessionKey, toolName, params, result, err)

	// Add graph relationships.
	if err != nil {
		e.recordErrorGraph(ctx, sessionKey, toolName, err)
	} else {
		e.propagateSuccess(ctx, toolName)
	}
}

// recordErrorGraph adds error-related triples to the knowledge graph.
func (e *GraphEngine) recordErrorGraph(ctx context.Context, sessionKey, toolName string, err error) {
	pattern := extractErrorPattern(err)
	errorNode := fmt.Sprintf("error:%s:%s", toolName, sanitizeForNode(pattern))
	sessionNode := fmt.Sprintf("session:%s", sessionKey)
	toolNode := fmt.Sprintf("tool:%s", toolName)

	triples := []graph.Triple{
		{
			Subject:   errorNode,
			Predicate: graph.CausedBy,
			Object:    toolNode,
		},
		{
			Subject:   errorNode,
			Predicate: graph.InSession,
			Object:    sessionNode,
		},
	}

	// Find similar errors from other tools.
	if e.graphStore != nil {
		existing, searchErr := e.graphStore.QueryBySubjectPredicate(ctx, errorNode, graph.SimilarTo)
		if searchErr == nil && len(existing) == 0 {
			// Search for similar patterns via knowledge store.
			entities, searchErr := e.store.SearchLearningEntities(ctx, pattern, 5)
			if searchErr == nil {
				for _, entity := range entities {
					if entity.ErrorPattern != "" && entity.ErrorPattern != pattern {
						similarNode := fmt.Sprintf("error:%s", sanitizeForNode(entity.ErrorPattern))
						triples = append(triples, graph.Triple{
							Subject:   errorNode,
							Predicate: graph.SimilarTo,
							Object:    similarNode,
						})
					}
				}
			}
		}
	}

	if e.graphCallback != nil {
		e.graphCallback(triples)
	} else if e.graphStore != nil {
		if addErr := e.graphStore.AddTriples(ctx, triples); addErr != nil {
			e.logger.Warnw("add error graph triples", "error", addErr)
		}
	}
}

// propagateSuccess boosts confidence of similar learnings when a tool succeeds.
func (e *GraphEngine) propagateSuccess(ctx context.Context, toolName string) {
	if e.graphStore == nil {
		return
	}

	toolNode := fmt.Sprintf("tool:%s", toolName)

	// Find errors caused by this tool.
	triples, err := e.graphStore.QueryByObject(ctx, toolNode)
	if err != nil {
		e.logger.Debugw("query tool errors", "error", err)
		return
	}

	for _, t := range triples {
		if t.Predicate != graph.CausedBy {
			continue
		}

		// Find similar errors and propagate confidence.
		similar, searchErr := e.graphStore.QueryBySubjectPredicate(ctx, t.Subject, graph.SimilarTo)
		if searchErr != nil {
			continue
		}

		for _, s := range similar {
			// Find learning entries for the similar error and boost their confidence.
			entities, searchErr := e.store.SearchLearningEntities(ctx, s.Object, 3)
			if searchErr != nil {
				continue
			}
			for _, entity := range entities {
				// Apply propagation rate as fractional confidence boost (e.g., 0.1 * 0.3 = 0.03).
				confidenceBoost := 0.1 * e.propagation
				if boostErr := e.store.BoostLearningConfidence(ctx, entity.ID, 1, confidenceBoost); boostErr != nil {
					e.logger.Debugw("propagate confidence", "error", boostErr)
				}
			}
		}
	}
}

// RecordFix records an error-fix relationship in the graph.
func (e *GraphEngine) RecordFix(ctx context.Context, errorPattern, fix, sessionKey string) {
	errorNode := fmt.Sprintf("error:%s", sanitizeForNode(errorPattern))
	fixNode := fmt.Sprintf("fix:%s", sanitizeForNode(fix))
	sessionNode := fmt.Sprintf("session:%s", sessionKey)

	triples := []graph.Triple{
		{
			Subject:   errorNode,
			Predicate: graph.ResolvedBy,
			Object:    fixNode,
		},
		{
			Subject:   fixNode,
			Predicate: graph.LearnedFrom,
			Object:    sessionNode,
		},
	}

	if e.graphCallback != nil {
		e.graphCallback(triples)
	} else if e.graphStore != nil {
		if addErr := e.graphStore.AddTriples(ctx, triples); addErr != nil {
			e.logger.Warnw("add fix graph triples", "error", addErr)
		}
	}
}

// sanitizeForNode makes a string safe for use as a graph node ID.
func sanitizeForNode(s string) string {
	if len(s) > 64 {
		s = s[:64]
	}
	// Replace spaces and special chars with underscores.
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == ':' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

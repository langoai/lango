package retrieval

import (
	"context"

	"github.com/langoai/lango/internal/knowledge"
)

// RetrievalAgent retrieves findings for specific context layers.
type RetrievalAgent interface {
	Name() string
	Layers() []knowledge.ContextLayer
	Search(ctx context.Context, query string, limit int) ([]Finding, error)
}

// FactSearchSource is the narrow interface for fact-based search operations.
// Satisfied by *knowledge.Store.
type FactSearchSource interface {
	SearchKnowledgeScored(ctx context.Context, query, category string, limit int) ([]knowledge.ScoredKnowledgeEntry, error)
	SearchLearningsScored(ctx context.Context, errorPattern, category string, limit int) ([]knowledge.ScoredLearningEntry, error)
	SearchExternalRefs(ctx context.Context, query string) ([]knowledge.ExternalRefEntry, error)
}

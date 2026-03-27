package retrieval

import "github.com/langoai/lango/internal/knowledge"

// Finding represents a single retrieved item with provenance metadata.
type Finding struct {
	Key          string                 // unique identifier
	Content      string                 // the retrieved text
	Score        float64                // normalized: higher = better
	Category     string                 // knowledge category
	SearchSource string                 // origin: "fts5", "like"
	Agent        string                 // producing agent name
	Layer        knowledge.ContextLayer // target context layer
}

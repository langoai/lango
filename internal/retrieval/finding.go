package retrieval

import (
	"time"

	"github.com/langoai/lango/internal/knowledge"
)

// Finding represents a single retrieved item with provenance metadata.
type Finding struct {
	Key          string                 // unique identifier
	Content      string                 // the retrieved text
	Score        float64                // normalized: higher = better
	Category     string                 // knowledge category
	SearchSource string                 // origin: "fts5", "like", "vector", "temporal"
	Agent        string                 // producing agent name
	Layer        knowledge.ContextLayer // target context layer

	// Provenance metadata for evidence-based merge.
	// Zero values mean "no provenance available" (e.g., ContextSearchAgent).
	Source    string    // authorship: "knowledge", "proactive_librarian", etc.
	Tags     []string  // includes "temporal:evergreen", "temporal:current_state"
	Version  int       // version chain position (higher supersedes lower)
	UpdatedAt time.Time // last update timestamp
}

package retrieval

import (
	"time"

	"github.com/langoai/lango/internal/knowledge"
)

// Finding represents a single retrieved item with provenance metadata.
type Finding struct {
	Key      string  // unique identifier
	Content  string  // the retrieved text
	Score    float64 // normalized: higher = better
	Category string  // knowledge category

	// SearchSource is the retrieval METHOD used to find this item.
	// Values: "fts5", "like", "vector", "temporal".
	// NOT to be confused with Source (authorship origin below).
	SearchSource string

	Agent string                 // producing agent name
	Layer knowledge.ContextLayer // target context layer

	// Provenance metadata for evidence-based merge.
	// Zero values mean "no provenance available".

	// Source is the AUTHORSHIP origin that produced this knowledge.
	// Values: "knowledge", "proactive_librarian", "session_learning",
	// "conversation_analysis", "memory", "learning".
	Source    string
	Tags      []string  // includes "temporal:evergreen", "temporal:current_state"
	Version   int       // version chain position (higher supersedes lower)
	UpdatedAt time.Time // last update timestamp
}

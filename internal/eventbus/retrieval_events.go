package eventbus

import "time"

// Event name constant for retrieval domain events.
const EventContextInjected = "context.injected"

// ContextInjectedEvent is published after context assembly in GenerateContent.
// It tracks which knowledge items were injected into the LLM system prompt for
// a given turn. Items contains only structured knowledge results (from
// ContextRetriever). RAG/memory/runSummary are represented as aggregate token
// counts only — no item-level decomposition.
type ContextInjectedEvent struct {
	TurnID           string // from session.TurnIDFromContext; "" if not in turn runner
	SessionKey       string
	Query            string // raw user query (processor must not log this — PII)
	Items            []ContextInjectedItem
	KnowledgeTokens  int
	RetrievedTokens  int
	MemoryTokens     int
	RunSummaryTokens int
	TotalTokens      int
	Timestamp        time.Time
}

// EventName implements Event.
func (e ContextInjectedEvent) EventName() string { return EventContextInjected }

// ContextInjectedItem represents a single knowledge item injected into context.
type ContextInjectedItem struct {
	Layer         string // human-readable layer name (from ContextLayer.String())
	Key           string
	Score         float64 // normalized: higher = better
	Source        string  // search source: "fts5", "like"
	Category      string
	TokenEstimate int
}

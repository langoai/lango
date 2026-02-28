package eventbus

// ContentSavedEvent is published when knowledge or memory content is saved.
// Replaces: SetEmbedCallback, SetGraphCallback on knowledge and memory stores.
type ContentSavedEvent struct {
	ID         string
	Collection string
	Content    string
	Metadata   map[string]string
	Source     string // "knowledge" or "memory"
}

// EventName implements Event.
func (e ContentSavedEvent) EventName() string { return "content.saved" }

// TriplesExtractedEvent is published when graph triples are extracted.
// Replaces: SetGraphCallback on learning engines and analyzers.
type TriplesExtractedEvent struct {
	Triples []Triple
	Source  string // e.g. "learning", "analysis", "librarian"
}

// EventName implements Event.
func (e TriplesExtractedEvent) EventName() string { return "triples.extracted" }

// Triple mirrors graph.Triple to avoid an import dependency on the graph
// package, keeping the eventbus package dependency-free.
type Triple struct {
	Subject   string
	Predicate string
	Object    string
	Metadata  map[string]string
}

// TurnCompletedEvent is published when a gateway turn completes.
// Replaces: Gateway.OnTurnComplete callbacks.
type TurnCompletedEvent struct {
	SessionKey string
}

// EventName implements Event.
func (e TurnCompletedEvent) EventName() string { return "turn.completed" }

// ReputationChangedEvent is published when a peer's reputation changes.
// Replaces: reputation.Store.SetOnChangeCallback.
type ReputationChangedEvent struct {
	PeerDID  string
	NewScore float64
}

// EventName implements Event.
func (e ReputationChangedEvent) EventName() string { return "reputation.changed" }

// MemoryGraphEvent is published when memory graph hooks fire.
// Replaces: memory.Store.SetGraphHooks.
type MemoryGraphEvent struct {
	Triples    []Triple
	SessionKey string
	Type       string // "observation" or "reflection"
}

// EventName implements Event.
func (e MemoryGraphEvent) EventName() string { return "memory.graph" }

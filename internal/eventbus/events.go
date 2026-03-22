package eventbus

// ContentSavedEvent is published when knowledge or memory content is saved.
// Replaces: SetEmbedCallback on knowledge and memory stores.
// Graph wiring subscribes only when NeedsGraph is true, preserving the original
// behavior where updates and learning saves skipped graph processing.
type ContentSavedEvent struct {
	ID         string
	Collection string
	Content    string
	Metadata   map[string]string
	Source     string // "knowledge" or "memory"
	IsNew      bool   // true for first-time creation, false for updates
	NeedsGraph bool   // true when graph triple extraction should also run
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

// ToolExecutionPaidEvent is published after a paid tool execution succeeds.
// The settlement service subscribes to this event to initiate on-chain settlement.
type ToolExecutionPaidEvent struct {
	PeerDID      string
	ToolName     string
	Auth         interface{} // *eip3009.Authorization (interface to avoid import cycle)
	SettlementID string      // non-empty for post-pay deferred entries
}

// EventName implements Event.
func (e ToolExecutionPaidEvent) EventName() string { return "tool.execution.paid" }

// --- P2P agent pool and discovery events ---

// AgentDiscoveredEvent is published when a new remote agent is discovered.
type AgentDiscoveredEvent struct {
	DID          string
	Name         string
	Capabilities []string
}

// EventName implements Event.
func (e AgentDiscoveredEvent) EventName() string { return "agent.discovered" }

// TaskDelegatedEvent is published when a task is delegated to an agent.
type TaskDelegatedEvent struct {
	TeamID   string
	TaskID   string
	AgentDID string
}

// EventName implements Event.
func (e TaskDelegatedEvent) EventName() string { return "task.delegated" }

// TaskCompletedEvent is published when a delegated task completes successfully.
type TaskCompletedEvent struct {
	TeamID     string
	TaskID     string
	AgentDID   string
	Success    bool
	DurationMs int64
}

// EventName implements Event.
func (e TaskCompletedEvent) EventName() string { return "task.completed" }

// TaskFailedEvent is published when a delegated task fails.
type TaskFailedEvent struct {
	TeamID   string
	TaskID   string
	AgentDID string
	Error    string
}

// EventName implements Event.
func (e TaskFailedEvent) EventName() string { return "task.failed" }

// PaymentNegotiatedEvent is published when payment terms are agreed.
type PaymentNegotiatedEvent struct {
	TeamID   string
	AgentDID string
	Mode     string
	Price    float64
}

// EventName implements Event.
func (e PaymentNegotiatedEvent) EventName() string { return "payment.negotiated" }

// PaymentSettledEvent is published when a payment is settled on-chain.
type PaymentSettledEvent struct {
	TeamID   string
	AgentDID string
	Amount   float64
	TxHash   string
}

// EventName implements Event.
func (e PaymentSettledEvent) EventName() string { return "payment.settled" }

// TrustUpdatedEvent is published when an agent's trust score changes.
type TrustUpdatedEvent struct {
	AgentDID string
	OldScore float64
	NewScore float64
}

// EventName implements Event.
func (e TrustUpdatedEvent) EventName() string { return "trust.updated" }

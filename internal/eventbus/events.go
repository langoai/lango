package eventbus

import "time"

// Event name constants for core domain events.
const (
	EventContentSaved      = "content.saved"
	EventTriplesExtracted  = "triples.extracted"
	EventTurnCompleted     = "turn.completed"
	EventReputationChanged = "reputation.changed"
	EventMemoryGraph       = "memory.graph"
	EventToolExecutionPaid = "tool.execution.paid"
	EventAgentDiscovered   = "agent.discovered"
	EventTaskDelegated     = "task.delegated"
	EventTaskCompleted     = "task.completed"
	EventTaskFailed        = "task.failed"
	EventPaymentNegotiated = "payment.negotiated"
	EventPaymentSettled    = "payment.settled"
	EventTrustUpdated      = "trust.updated"
	EventSchemaExchanged   = "schema.exchanged"
	EventPolicyDecision    = "policy.decision"
	EventAlertTriggered    = "alert.triggered"
)

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
	Version    int    // Knowledge version; 0 for non-versioned collections
}

// EventName implements Event.
func (e ContentSavedEvent) EventName() string { return EventContentSaved }

// TriplesExtractedEvent is published when graph triples are extracted.
// Replaces: SetGraphCallback on learning engines and analyzers.
type TriplesExtractedEvent struct {
	Triples []Triple
	Source  string // e.g. "learning", "analysis", "librarian"
}

// EventName implements Event.
func (e TriplesExtractedEvent) EventName() string { return EventTriplesExtracted }

// Triple mirrors graph.Triple to avoid an import dependency on the graph
// package, keeping the eventbus package dependency-free.
type Triple struct {
	Subject     string
	Predicate   string
	Object      string
	SubjectType string
	ObjectType  string
	Metadata    map[string]string
}

// TurnCompletedEvent is published when a gateway turn completes.
// Replaces: Gateway.OnTurnComplete callbacks.
type TurnCompletedEvent struct {
	SessionKey string
}

// EventName implements Event.
func (e TurnCompletedEvent) EventName() string { return EventTurnCompleted }

// ReputationChangedEvent is published when a peer's reputation changes.
// Replaces: reputation.Store.SetOnChangeCallback.
type ReputationChangedEvent struct {
	PeerDID  string
	NewScore float64
}

// EventName implements Event.
func (e ReputationChangedEvent) EventName() string { return EventReputationChanged }

// MemoryGraphEvent is published when memory graph hooks fire.
// Replaces: memory.Store.SetGraphHooks.
type MemoryGraphEvent struct {
	Triples    []Triple
	SessionKey string
	Type       string // "observation" or "reflection"
}

// EventName implements Event.
func (e MemoryGraphEvent) EventName() string { return EventMemoryGraph }

// ToolExecutionPaidEvent is published after a paid tool execution succeeds.
// The settlement service subscribes to this event to initiate on-chain settlement.
type ToolExecutionPaidEvent struct {
	PeerDID      string
	ToolName     string
	Auth         interface{} // *eip3009.Authorization (interface to avoid import cycle)
	SettlementID string      // non-empty for post-pay deferred entries
}

// EventName implements Event.
func (e ToolExecutionPaidEvent) EventName() string { return EventToolExecutionPaid }

// --- P2P agent pool and discovery events ---

// AgentDiscoveredEvent is published when a new remote agent is discovered.
type AgentDiscoveredEvent struct {
	DID          string
	Name         string
	Capabilities []string
}

// EventName implements Event.
func (e AgentDiscoveredEvent) EventName() string { return EventAgentDiscovered }

// TaskDelegatedEvent is published when a task is delegated to an agent.
type TaskDelegatedEvent struct {
	TeamID   string
	TaskID   string
	AgentDID string
}

// EventName implements Event.
func (e TaskDelegatedEvent) EventName() string { return EventTaskDelegated }

// TaskCompletedEvent is published when a delegated task completes successfully.
type TaskCompletedEvent struct {
	TeamID     string
	TaskID     string
	AgentDID   string
	Success    bool
	DurationMs int64
}

// EventName implements Event.
func (e TaskCompletedEvent) EventName() string { return EventTaskCompleted }

// TaskFailedEvent is published when a delegated task fails.
type TaskFailedEvent struct {
	TeamID   string
	TaskID   string
	AgentDID string
	Error    string
}

// EventName implements Event.
func (e TaskFailedEvent) EventName() string { return EventTaskFailed }

// PaymentNegotiatedEvent is published when payment terms are agreed.
type PaymentNegotiatedEvent struct {
	TeamID   string
	AgentDID string
	Mode     string
	Price    float64
}

// EventName implements Event.
func (e PaymentNegotiatedEvent) EventName() string { return EventPaymentNegotiated }

// PaymentSettledEvent is published when a payment is settled on-chain.
type PaymentSettledEvent struct {
	TeamID   string
	AgentDID string
	Amount   float64
	TxHash   string
}

// EventName implements Event.
func (e PaymentSettledEvent) EventName() string { return EventPaymentSettled }

// TrustUpdatedEvent is published when an agent's trust score changes.
type TrustUpdatedEvent struct {
	AgentDID string
	OldScore float64
	NewScore float64
}

// EventName implements Event.
func (e TrustUpdatedEvent) EventName() string { return EventTrustUpdated }

// SchemaExchangeEvent is published after a P2P ontology schema exchange.
type SchemaExchangeEvent struct {
	PeerDID    string // remote peer DID
	Direction  string // "export" or "import"
	TypeCount  int    // number of types exchanged
	PredCount  int    // number of predicates exchanged
	ImportMode string // import mode used (empty for export)
}

// EventName implements Event.
func (e SchemaExchangeEvent) EventName() string { return EventSchemaExchanged }

// PolicyDecisionEvent is published when the exec policy evaluator makes
// an observe or block decision. Allow verdicts are not published.
type PolicyDecisionEvent struct {
	Command    string
	Unwrapped  string
	Verdict    string // "allow", "observe", "block"
	Reason     string // machine-readable reason code
	Message    string
	SessionKey string
	AgentName  string
}

// EventName implements Event.
func (e PolicyDecisionEvent) EventName() string { return EventPolicyDecision }

// AlertEvent is published when an operational alert condition is detected.
type AlertEvent struct {
	Type       string                 // "policy_block_rate", "recovery_retries", "circuit_breaker", "config_drift"
	Severity   string                 // "warning", "critical"
	Message    string
	Details    map[string]interface{}
	SessionKey string
	Timestamp  time.Time
}

// EventName implements Event.
func (e AlertEvent) EventName() string { return EventAlertTriggered }

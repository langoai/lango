## Purpose

Capability spec for retrieval-feedback-observability. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: TurnID context propagation
The system SHALL propagate TurnID through `context.Context` using `session.WithTurnID(ctx, turnID)` and `session.TurnIDFromContext(ctx)`. TurnRunner SHALL set TurnID in context before calling the executor. When TurnID is absent (e.g., direct calls without TurnRunner), `TurnIDFromContext` SHALL return an empty string.

#### Scenario: TurnID round-trip
- **WHEN** `WithTurnID(ctx, "turn-abc-123")` is called
- **THEN** `TurnIDFromContext(ctx)` SHALL return `"turn-abc-123"`

#### Scenario: TurnID absent
- **WHEN** `TurnIDFromContext` is called on a context without TurnID
- **THEN** it SHALL return an empty string

#### Scenario: TurnRunner sets TurnID
- **WHEN** `TurnRunner.Run()` executes a turn
- **THEN** the context passed to the executor SHALL contain the turn's traceID via `WithTurnID`

### Requirement: ContextInjectedEvent
The system SHALL publish a `ContextInjectedEvent` via the event bus after context assembly in `GenerateContent`, before the LLM call. The event SHALL contain: TurnID (may be empty), SessionKey, Query, structured knowledge Items, per-section token estimates (KnowledgeTokens, RetrievedTokens, MemoryTokens, RunSummaryTokens, TotalTokens), and Timestamp. Items SHALL contain only knowledge `ContextRetriever` structured results. Retrieved context/memory/runSummary SHALL be represented as aggregate token counts only.

#### Scenario: Event published after context assembly
- **WHEN** `GenerateContent` completes context assembly and the event bus is set
- **THEN** a `ContextInjectedEvent` SHALL be published with all retrieved knowledge items, their scores, sources, and per-section token estimates

### Requirement: ContextInjectedItem structure
Each `ContextInjectedItem` SHALL contain: Layer (human-readable string from `ContextLayer.String()`), Key, Score (float64, higher=better), Source (search source: "fts5", "like"), Category, and TokenEstimate. The Layer field SHALL be a string (not `ContextLayer`) to keep the eventbus package dependency-free.

#### Scenario: Item fields populated from knowledge retrieval
- **WHEN** a knowledge item with key "deploy-config", score 0.9, source "fts5", category "fact" is retrieved
- **THEN** the corresponding `ContextInjectedItem` SHALL have Layer="user_knowledge", Key="deploy-config", Score=0.9, Source="fts5", Category="fact", and TokenEstimate equal to the estimated token count of the item content

### Requirement: FeedbackProcessor
The system SHALL provide a `FeedbackProcessor` that subscribes to `ContextInjectedEvent` and performs structured logging. The processor SHALL log: session_key, query_length, knowledge_items count, per-section token counts, total_tokens, layer_distribution (map of layer to count), and source_distribution (map of source to count). The processor SHALL include turn_id in logs only when non-empty. The processor SHALL NOT log raw query text (PII consideration). The processor SHALL NOT modify any stored data (no score adjustment, no use-count increment).

#### Scenario: Structured logging with TurnID
- **WHEN** `FeedbackProcessor` receives a `ContextInjectedEvent` with TurnID "turn-123"
- **THEN** the log entry SHALL include "turn_id"="turn-123", "knowledge_items" count, "retrieved_tokens", "total_tokens", "layer_distribution", and "source_distribution"

### Requirement: ContextLayer.String()
The `ContextLayer` type SHALL provide a `String()` method returning human-readable snake_case names: "tool_registry", "user_knowledge", "skill_patterns", "external_knowledge", "agent_learnings", "runtime_context", "observations", "reflections", "pending_inquiries". Unknown layers SHALL return "layer_N" where N is the integer value.

#### Scenario: Known layer string representation
- **WHEN** `LayerUserKnowledge.String()` is called
- **THEN** it SHALL return `"user_knowledge"`

#### Scenario: Unknown layer string representation
- **WHEN** `ContextLayer(99).String()` is called
- **THEN** it SHALL return `"layer_99"`


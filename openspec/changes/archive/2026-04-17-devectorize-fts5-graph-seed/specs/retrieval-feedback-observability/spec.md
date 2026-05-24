## MODIFIED Requirements

### Requirement: ContextInjectedEvent
The system SHALL publish a `ContextInjectedEvent` via the event bus after context assembly in `GenerateContent`, before the LLM call. The event SHALL contain: TurnID (may be empty), SessionKey, Query, structured knowledge Items, per-section token estimates (KnowledgeTokens, RetrievedTokens, MemoryTokens, RunSummaryTokens, TotalTokens), and Timestamp. Items SHALL contain only knowledge `ContextRetriever` structured results. Retrieved context/memory/runSummary SHALL be represented as aggregate token counts only.

#### Scenario: Event published after context assembly
- **WHEN** `GenerateContent` completes context assembly and the event bus is set
- **THEN** a `ContextInjectedEvent` SHALL be published with all retrieved knowledge items, their scores, sources, and per-section token estimates

### Requirement: FeedbackProcessor
The system SHALL provide a `FeedbackProcessor` that subscribes to `ContextInjectedEvent` and performs structured logging. The processor SHALL log: session_key, query_length, knowledge_items count, per-section token counts, total_tokens, layer_distribution (map of layer to count), and source_distribution (map of source to count). The processor SHALL include turn_id in logs only when non-empty. The processor SHALL NOT log raw query text (PII consideration). The processor SHALL NOT modify any stored data (no score adjustment, no use-count increment).

#### Scenario: Structured logging with TurnID
- **WHEN** `FeedbackProcessor` receives a `ContextInjectedEvent` with TurnID "turn-123"
- **THEN** the log entry SHALL include "turn_id"="turn-123", "knowledge_items" count, "retrieved_tokens", "total_tokens", "layer_distribution", and "source_distribution"

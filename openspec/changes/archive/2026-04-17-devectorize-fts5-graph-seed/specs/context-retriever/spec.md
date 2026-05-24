## MODIFIED Requirements

### Requirement: Context-Aware Model Adapter
The system SHALL wrap the ADK model adapter to transparently inject retrieved context and observational memory. The GenerateContent method SHALL execute knowledge retrieval, GraphRAG retrieval, and memory retrieval in parallel using errgroup. Each retrieval SHALL run as a separate goroutine with context cancellation propagation. Errors in individual retrievals SHALL be logged and treated as non-fatal (existing degradation pattern preserved).

#### Scenario: System prompt augmentation
- **WHEN** `GenerateContent` is called on the context-aware adapter
- **THEN** the system SHALL extract the user's latest message as query
- **AND** retrieve relevant context from all 6 layers (Runtime Context, Tool Registry, User Knowledge, Skill Patterns, External Knowledge, Agent Learnings)
- **AND** update the runtime adapter's session state before retrieval
- **AND** retrieve graph-enhanced retrieved context for the current query when GraphRAG is configured
- **AND** retrieve observations and reflections for the current session
- **AND** augment the system instruction with assembled context including observations
- **AND** forward the modified request to the underlying model adapter

#### Scenario: All three retrievals run concurrently
- **WHEN** GenerateContent is called with knowledge, GraphRAG, and memory providers configured
- **THEN** all three retrievals SHALL execute concurrently and their results combined after completion

#### Scenario: One retrieval fails
- **WHEN** knowledge retrieval fails but GraphRAG and memory succeed
- **THEN** the error SHALL be logged and the prompt SHALL include retrieved context and memory sections only

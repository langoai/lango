## Purpose

Child-session routing for isolated sub-agents. Prevents specialist raw turns from polluting parent session history while preserving same-run causal visibility and summary-based merge/discard behavior.
## Requirements
### Requirement: Child summary preserves typed incomplete cause
When an isolated child session ends without a visible assistant summary, the parent-visible note SHALL preserve the classified incomplete cause instead of using a generic placeholder.

#### Scenario: Empty-after-tool-use note is explicit
- **WHEN** an isolated specialist ends without visible assistant completion after tool activity
- **THEN** the parent-visible note SHALL mention `empty_after_tool_use`
- **AND** it SHALL NOT promote raw tool output to a success summary

## ADDED Requirements

### Requirement: ChildSession type
The `session` package SHALL define a `ChildSession` type with fields: ID, ParentID, AgentName, Config, CreatedAt, and Status. ChildSession SHALL support cross-turn isolation while allowing same-run causal visibility through the active parent session view.

#### Scenario: Child session reads parent history
- **WHEN** a ChildSession is created from a parent session
- **THEN** it SHALL be able to read the parent's message history

#### Scenario: Child session same-run overlay visibility
- **WHEN** an isolated child session appends events during an active run
- **THEN** those events SHALL be visible through the parent session's in-memory event stream for the remainder of that run
- **AND** they SHALL NOT be persisted as raw messages in the parent store

#### Scenario: Child session writes remain cross-turn isolated
- **WHEN** the next turn reloads the parent session from persistent storage
- **THEN** raw child events SHALL NOT appear in the parent session history

### Requirement: ChildSessionStore interface
The package SHALL define a `ChildSessionStore` interface with methods: ForkChild, MergeChild, DiscardChild.

#### Scenario: Fork creates isolated child
- **WHEN** ForkChild is called with a parent session ID
- **THEN** a new ChildSession SHALL be created with access to parent history

#### Scenario: Merge brings back summary only
- **WHEN** MergeChild is called on a completed child session
- **THEN** the parent persistent history SHALL receive only a summary outcome
- **AND** raw child events SHALL remain absent from the persisted parent history

#### Scenario: Discard leaves compact failure note only
- **WHEN** DiscardChild is called after a runtime failure with a discard reason
- **THEN** the child session data SHALL be cleaned up without affecting the raw parent history
- **AND** the parent persistent history MAY receive only a compact root-authored failure note

### Requirement: StructuredSummarizer
The `adk` package SHALL provide a `StructuredSummarizer` that extracts the last assistant response from a child session as the merge result. This SHALL be the default summarizer (zero LLM cost).

#### Scenario: Extract last response
- **WHEN** StructuredSummarizer processes a child session with multiple messages
- **THEN** it SHALL return only the content of the last assistant message

### Requirement: ChildSessionServiceAdapter
The `adk` package SHALL provide a `ChildSessionServiceAdapter` that bridges the ChildSessionStore with ADK's session management for sub-agent isolation.

#### Scenario: Sub-agent gets isolated session
- **WHEN** a sub-agent is invoked with session isolation enabled
- **THEN** it SHALL receive a forked child session with parent context but isolated writes

### Requirement: Isolation activation is independent of provenance observers
Enabling session isolation for an agent SHALL activate child-session routing even when provenance/session-tree observers are not configured.

#### Scenario: Isolated specialist without provenance still forks child state
- **WHEN** `WithAgentIsolatedAgents()` marks `vault` as isolated and no provenance child-lifecycle observer is installed
- **THEN** `vault` turns SHALL still be routed through a child session
- **AND** child-session merge/discard behavior SHALL remain active

### Requirement: Raw isolated turns never persist to the parent session store
Raw assistant/tool messages authored by isolated agents SHALL NOT be persisted into the parent session store under any runtime configuration. Only merged summaries or compact discard notes may persist in the parent history.

#### Scenario: Successful isolated run persists summary only
- **WHEN** an isolated specialist completes successfully
- **THEN** the parent session store SHALL persist only the merged summary authored as the root/orchestrator agent
- **AND** raw isolated assistant/tool turns SHALL remain absent from persisted parent history

#### Scenario: Failed isolated run persists discard note only
- **WHEN** an isolated specialist run is discarded after failure
- **THEN** the parent session store SHALL persist only a compact discard note
- **AND** raw isolated assistant/tool turns SHALL remain absent from persisted parent history

### Requirement: Discard notes include classified failure reason
When an isolated child session is discarded, the persisted parent note SHALL include the runtime failure classification while continuing to exclude raw child history.

#### Scenario: Loop failure discard note
- **WHEN** an isolated specialist run is discarded because of repeated identical tool calls
- **THEN** the parent discard note SHALL include the classification `loop_detected`
- **AND** SHALL NOT embed raw child messages or tool payloads

#### Scenario: Empty after tool use discard note
- **WHEN** an isolated specialist run is discarded because tool work completed without visible synthesis
- **THEN** the parent discard note SHALL include the classification `empty_after_tool_use`
- **AND** SHALL NOT embed raw child messages or tool payloads

### Requirement: Streaming failure paths discard isolated child sessions
Streaming runtime failure paths SHALL discard active isolated child sessions using the same classified discard behavior as collection-based failure paths.

#### Scenario: Iterator error discards active isolated child
- **WHEN** `RunStreamingDetailed` receives an iterator error while an isolated specialist child session is active
- **THEN** the runtime SHALL discard the active child session with the classified discard reason before returning the error
- **AND** the next retry or turn SHALL not observe the stale child overlay

#### Scenario: Streaming discard keeps parent persistence compact
- **WHEN** a streaming failure discards an isolated child session
- **THEN** persisted parent history SHALL include at most the compact root-authored discard note
- **AND** raw isolated specialist assistant and tool messages SHALL remain absent from the parent store

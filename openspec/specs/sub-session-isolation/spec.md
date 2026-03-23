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

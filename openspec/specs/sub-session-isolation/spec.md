## ADDED Requirements

### Requirement: ChildSession type
The `session` package SHALL define a `ChildSession` type with fields: ID, ParentID, AgentName, Config, CreatedAt, and Status. ChildSession SHALL support "read parent, write child" isolation.

#### Scenario: Child session reads parent history
- **WHEN** a ChildSession is created from a parent session
- **THEN** it SHALL be able to read the parent's message history

#### Scenario: Child session writes are isolated
- **WHEN** a ChildSession appends events
- **THEN** the events SHALL NOT appear in the parent session's history

### Requirement: ChildSessionStore interface
The package SHALL define a `ChildSessionStore` interface with methods: ForkChild, MergeChild, DiscardChild.

#### Scenario: Fork creates isolated child
- **WHEN** ForkChild is called with a parent session ID
- **THEN** a new ChildSession SHALL be created with access to parent history

#### Scenario: Merge brings results back
- **WHEN** MergeChild is called on a completed child session
- **THEN** the child's result (via summarizer) SHALL be appended to the parent session

#### Scenario: Discard removes child
- **WHEN** DiscardChild is called
- **THEN** the child session data SHALL be cleaned up without affecting the parent

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

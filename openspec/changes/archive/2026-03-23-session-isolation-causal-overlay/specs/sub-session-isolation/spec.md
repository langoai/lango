## MODIFIED Requirements

### Requirement: ChildSession type
The `session` package SHALL define a `ChildSession` type with fields: ID, ParentID, AgentName, Config, CreatedAt, and Status. ChildSession SHALL support cross-turn isolation while allowing same-run causal visibility through the active parent session view.

#### Scenario: Child session same-run overlay visibility
- **WHEN** an isolated child session appends events during an active run
- **THEN** those events SHALL be visible through the parent session's in-memory event stream for the remainder of that run
- **AND** they SHALL NOT be persisted as raw messages in the parent store

#### Scenario: Child session writes remain cross-turn isolated
- **WHEN** the next turn reloads the parent session from persistent storage
- **THEN** raw child events SHALL NOT appear in the parent session history

### Requirement: ChildSessionStore interface
The package SHALL define a `ChildSessionStore` interface with methods: ForkChild, MergeChild, DiscardChild.

#### Scenario: Merge brings back summary only
- **WHEN** an isolated child session is merged successfully
- **THEN** the parent persistent history SHALL receive only a summary outcome
- **AND** raw child events SHALL remain absent from the persisted parent history

#### Scenario: Discard leaves compact failure note only
- **WHEN** an isolated child session is discarded with a runtime failure reason
- **THEN** the parent persistent history SHALL receive only a compact root-authored failure note
- **AND** raw child events SHALL remain absent from the persisted parent history
